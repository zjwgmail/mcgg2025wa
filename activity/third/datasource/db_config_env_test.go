package datasource

import (
	"go-fission-activity/activity/web/middleware/logTracing"
	"strings"
	"testing"
)

func TestReplaceEnv(t *testing.T) {
	DB_URL := "jdbc:postgresql://172.16.100.229:5432"
	var dbUrlPrefix = "jdbc:postgresql://"

	if strings.Contains(DB_URL, dbUrlPrefix) {
		dbSplitUrl := strings.Split(strings.Replace(DB_URL, "jdbc:postgresql://", "", 1), ":")
		logTracing.LogPrintfP("host:%s, port:%s", dbSplitUrl[0], dbSplitUrl[1])
	}
}
