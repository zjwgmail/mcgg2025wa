package constant

const (
	ThreeCdk = "ThreeCdk" // 3人
	FiveCdk  = "FiveCdk"  // 5人
	EightCdk = "EightCdk" // 8人
	FreeCdk  = "FreeCdk"  // 8人
)

func GetAllCdkType() []string {
	return []string{FreeCdk, ThreeCdk, FiveCdk, EightCdk}
}

func ContainsCdkType(cdkType string) bool {
	allCdkType := GetAllCdkType()
	for _, v := range allCdkType {
		if v == cdkType {
			return true
		}
	}
	return false
}
