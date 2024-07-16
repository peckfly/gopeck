package enums

const (
	Rps         StressType = 1
	Concurrency StressType = 2

	DOING TaskStatus = 1
	DONE  TaskStatus = 2
	STOP  TaskStatus = 3

	Constants StressModeType = 1
	Step      StressModeType = 2

	RedisSingle  = 0
	RedisCluster = 1
)

type (
	StressType     int
	StressModeType int
	TaskStatus     int
)

var (
	systemSupportStressType = []StressType{
		Rps,
		Concurrency,
	}

	systemSupportStressModeType = []StressModeType{
		Constants,
		Step,
	}
)

func SupportStressType(stressType int) bool {
	for _, t := range systemSupportStressType {
		if t == StressType(stressType) {
			return true
		}
	}
	return false
}

func SupportStressModeType(stressModeType int) bool {
	for _, t := range systemSupportStressModeType {
		if t == StressModeType(stressModeType) {
			return true
		}
	}
	return false
}
