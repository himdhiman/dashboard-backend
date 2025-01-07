package scheduler

type ErrorCode string

const (
	ErrCodeInvalidConfig ErrorCode = "INVALID_CONFIG"
	ErrCodeJobNotFound   ErrorCode = "JOB_NOT_FOUND"
	ErrCodeJobFailed     ErrorCode = "JOB_FAILED"
)

type SchedulerError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *SchedulerError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}
