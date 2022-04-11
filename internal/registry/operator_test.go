package registry

import (
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	name := "zoperator.v0.3.6"
	ocpversion := "4.6"
	path, _ := os.Getwd()
	log.Info(path)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	loadOperatorsCatalog(path + "/../")
	ans := IsOperatorCertified(name, ocpversion)
	fmt.Println(ans)
	assert.Equal(t, ans, true)
}
