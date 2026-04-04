package app

import (
	"context"
	"time"

	"server-monitor/internal/collector"
	"server-monitor/internal/config"
	"server-monitor/internal/logx"
	"server-monitor/internal/model"
	agentruntime "server-monitor/internal/runtime"
	"server-monitor/internal/sender"
)

const successLogEveryLoops = 10

func Run(ctx context.Context, cfg config.Config) {
	staticMetrics := collector.CollectStaticMetrics(cfg.ServerID)
	metricsSender := sender.New(cfg.HTTPTimeoutSeconds, cfg.FallbackIntervalSeconds)
	state := agentruntime.NewState(time.Now())

	initialInterval, err := metricsSender.SendMetrics(cfg.APIURL, cfg.APIKey, staticMetrics)
	if err != nil {
		state.MarkFailure(err)
		logx.Error("Initial static metrics send failed: %v", err)
		initialInterval = cfg.FallbackIntervalSeconds
	} else {
		state.MarkSuccess(time.Now())
		logx.Info("Initial static metrics send succeeded. Interval: %d seconds", initialInterval)
	}

	interval := initialInterval
	loopCount := 0
	cachedProcesses := []model.ProcessMetrics{}
	cachedSupplemental := model.SupplementalMetrics{}

	for {
		select {
		case <-ctx.Done():
			logx.Info(
				"Warden shutting down. Started at: %s, successful sends: %d, last success: %s, consecutive failures: %d",
				state.StartedAt.Format(time.RFC3339),
				state.SuccessfulSends,
				formatOptionalTime(state.LastSuccessAt),
				state.ConsecutiveFailures,
			)
			return
		default:
		}

		loopCount++
		if shouldCollectProcesses(loopCount, cfg) {
			cachedProcesses = collector.CollectProcessMetrics(cfg.MaxProcesses)
		}
		if shouldCollectSupplementalMetrics(loopCount, cfg) {
			cachedSupplemental = collector.CollectSupplementalMetrics()
		}

		dynamicMetrics := collector.CollectDynamicMetrics(cachedProcesses, cachedSupplemental)
		dynamicMetrics.ServerID = staticMetrics.ServerID
		dynamicMetrics.AppVersion = staticMetrics.AppVersion
		dynamicMetrics.Agent = staticMetrics.Agent

		newInterval, err := metricsSender.SendMetrics(cfg.APIURL, cfg.APIKey, dynamicMetrics)
		if err != nil {
			state.MarkFailure(err)
			logx.Error(
				"Dynamic metrics send failed: %v | consecutive failures: %d | last success: %s",
				err,
				state.ConsecutiveFailures,
				formatOptionalTime(state.LastSuccessAt),
			)
		} else {
			state.MarkSuccess(time.Now())
			previousInterval := interval
			interval = newInterval
			if interval != previousInterval {
				logx.Info("Dynamic metrics send succeeded. Interval updated: %d -> %d seconds", previousInterval, interval)
			} else if loopCount%successLogEveryLoops == 0 {
				logx.Info("Dynamic metrics send succeeded. Interval: %d seconds | successful sends: %d", interval, state.SuccessfulSends)
			}
		}

		timer := time.NewTimer(time.Duration(interval) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			logx.Info(
				"Warden shutting down. Started at: %s, successful sends: %d, last success: %s, consecutive failures: %d",
				state.StartedAt.Format(time.RFC3339),
				state.SuccessfulSends,
				formatOptionalTime(state.LastSuccessAt),
				state.ConsecutiveFailures,
			)
			return
		case <-timer.C:
		}
	}
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return "none"
	}

	return value.Format(time.RFC3339)
}

func shouldCollectProcesses(loopCount int, cfg config.Config) bool {
	if cfg.MaxProcesses == 0 {
		return false
	}

	if loopCount == 1 {
		return true
	}

	return loopCount%cfg.ProcessCollectionIntervalMultiple == 0
}

func shouldCollectSupplementalMetrics(loopCount int, cfg config.Config) bool {
	if loopCount == 1 {
		return true
	}

	return loopCount%cfg.ProcessCollectionIntervalMultiple == 0
}
