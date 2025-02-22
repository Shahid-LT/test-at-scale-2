package teststats

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/LambdaTest/synapse/config"
	"github.com/LambdaTest/synapse/pkg/core"
	"github.com/LambdaTest/synapse/pkg/global"
	"github.com/LambdaTest/synapse/pkg/lumber"
	"github.com/LambdaTest/synapse/pkg/procfs"
)

//ProcStats represents the process stats for a particular pid
type ProcStats struct {
	logger                       lumber.Logger
	httpClient                   http.Client
	ExecutionResultInputChannel  chan core.ExecutionResult
	wg                           sync.WaitGroup
	ExecutionResultOutputChannel chan core.ExecutionResult
}

// New returns instance of ProcStats
func New(cfg *config.NucleusConfig, logger lumber.Logger) (*ProcStats, error) {
	return &ProcStats{
		logger:                      logger,
		ExecutionResultInputChannel: make(chan core.ExecutionResult),
		httpClient: http.Client{
			Timeout: global.DefaultHTTPTimeout,
		},
		ExecutionResultOutputChannel: make(chan core.ExecutionResult),
	}, nil

}

// CaptureTestStats combines the ps stats for each test
func (s *ProcStats) CaptureTestStats(pid int32) error {
	ps, err := procfs.New(pid, global.SamplingTime, false)
	if err != nil {
		s.logger.Errorf("failed to find process stats with pid %d %v", pid, err)
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		processStats := ps.GetStatsInInterval()
		if len(processStats) == 0 {
			s.logger.Errorf("no process stats found with pid %d", pid)
		}
		select {
		case executionResult := <-s.ExecutionResultInputChannel:
			// Refactor the impl of below 2 functions using generics when Go 1.18 arrives
			// https://www.freecodecamp.org/news/generics-in-golang/
			s.appendStatsToTests(executionResult.TestPayload, processStats)
			s.appendStatsToTestSuites(executionResult.TestSuitePayload, processStats)

			s.ExecutionResultOutputChannel <- executionResult
		default:
			// Can reach here in 2 cases (ie `/results` API wasn't called):
			// 1. runner process exited with zero exit exitCode but no testFiles were run (changes in Readme.md etc)
			// 2. runner process exited with non-zero exitCode
			// In second case, non-zero exitCodes are already captured and sent as
			// "Task error" when updating task status to neuron in lifeycle.go
			s.logger.Warnf("No test results found, pid %d", pid)
			s.ExecutionResultOutputChannel <- core.ExecutionResult{}
		}
	}()

	return nil
}

// processStats is RecordTime sorted
func (s *ProcStats) getProcsForInterval(start, end time.Time, processStats []*procfs.Stats) []*procfs.Stats {
	n := len(processStats)
	left := sort.Search(n, func(i int) bool { return !processStats[i].RecordTime.Before(start) })
	right := sort.Search(n, func(i int) bool { return !processStats[i].RecordTime.Before(end) })

	if left <= right && 0 <= left && right <= n {
		return processStats[left:right]
	}
	// return empty slice
	return processStats[0:0]
}

func (s *ProcStats) appendStatsToTests(testResults []core.TestPayload, processStats []*procfs.Stats) {
	for r := 0; r < len(testResults); r++ {
		result := &testResults[r]
		// check if start time of test t(start) is not 0
		if !result.StartTime.IsZero() {
			// calculate end time of test t(end)
			result.EndTime = result.StartTime.Add(time.Duration(result.Duration) * time.Millisecond)
			for _, proc := range s.getProcsForInterval(result.StartTime, result.EndTime, processStats) {
				result.Stats = append(result.Stats, core.TestProcessStats{CPU: proc.CPUPercentage, Memory: proc.MemConsumed, RecordTime: proc.RecordTime})
			}
		}
	}
}

func (s *ProcStats) appendStatsToTestSuites(testSuiteResults []core.TestSuitePayload, processStats []*procfs.Stats) {
	for r := 0; r < len(testSuiteResults); r++ {
		result := &testSuiteResults[r]
		// check if start time of test suite ts(start) is not 0
		if !result.StartTime.IsZero() {
			// calculate end time of test suite ts(end)
			result.EndTime = result.StartTime.Add(time.Duration(result.Duration) * time.Millisecond)
			for _, proc := range s.getProcsForInterval(result.StartTime, result.EndTime, processStats) {
				result.Stats = append(result.Stats, core.TestProcessStats{CPU: proc.CPUPercentage, Memory: proc.MemConsumed, RecordTime: proc.RecordTime})
			}
		}
	}
}
