package initConfig

const (
	ReFreeFirstHour = 22
	ReFreeNextHour  = 36
	CdkLimit        = 90
)

func GetReFreeFirstHour() int {
	return ReFreeFirstHour
}

func GetReFreeNextHour() int {
	return ReFreeNextHour
}
func GetCdkLimit() float64 {
	return CdkLimit
}

func IsConfigActivity() bool {
	// todo 如果不判断国码，则返回false
	return true
}
