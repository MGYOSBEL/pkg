package healthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Health struct {
	Details map[string]bool `json:"details"`
	Status  bool            `json:"status"`
}

type Checker interface {
	Check() (bool, error)
}

type HealthChecker struct {
	health   Health
	checkers map[string]Checker
	sugar    *zap.SugaredLogger
	endpoint string
}

func New(endpoint string, logger *zap.Logger) *HealthChecker {
	hc := &HealthChecker{
		health: Health{
			Details: make(map[string]bool),
			Status:  true,
		},
		checkers: make(map[string]Checker),
		sugar:    logger.Sugar(),
		endpoint: endpoint,
	}

	hc.RegisterGlobalStatus()

	return hc
}

func (hc *HealthChecker) Register(name string, ch Checker) {
	hc.checkers[name] = ch
	endpoint := fmt.Sprintf("/%s/%s", hc.endpoint, name)
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		health, _ := (ch).Check()
		res, _ := json.Marshal(health)
		if health == false {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = w.Write(res)
	})
	hc.sugar.Debugf("❤️ Registered dependency healthcheck in endpoint: \"%v\"", endpoint)
}

func (hc *HealthChecker) RegisterGlobalStatus() {
	endpoint := fmt.Sprintf("/%s", hc.endpoint)
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		hc.Status()
		res, _ := json.Marshal(hc.health)
		if hc.health.Status == false {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = w.Write(res)
	})
	hc.sugar.Infof("❤️ Global healthcheck registered in endpoint: \"%v\"", endpoint)
}

func (hc *HealthChecker) Status() {
	hc.health.Status = true
	for name, chk := range hc.checkers {
		healthy, err := chk.Check()
		if err != nil {
			hc.sugar.Errorf("🔴 Error checking %v", name)
			hc.health.Details[name] = false
		} else {
			hc.health.Details[name] = healthy
			hc.sugar.Debugf("❤️ %s checker returns %v status", name, healthy)
		}
		hc.health.Status = hc.health.Status && healthy
	}
	hc.sugar.Debugf("❤️ Global healty status is %v", hc.health.Status)
}
