package azure

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzure_GetChangesetChanges(t *testing.T) {
	conf := NewConfig()
	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

	azure := NewAzure(conf)
	azure.Connect()
	azure.TfvcClientConnection()
	projects, err := azure.ListOfProjects()
	assert.NoError(t, err)

	for _, p := range projects {
		changesets, _ := azure.GetChangesets(*p)
		fmt.Println("LAST", *(changesets)[1])
		for _, v := range changesets {
			fmt.Println(azure.GetChangesetChanges(v, *p))
		}
	}
}

func TestAzure_ChangedRows(t *testing.T) {
	conf := NewConfig()
	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

	azure := NewAzure(conf)
	azure.Connect()
	azure.TfvcClientConnection()

	//ссылки на файлы
	projectName := "Project2"
	currentFileUrl := "$/Project2/test.txt"
	previousFileUrl := "https://dev.azure.com/GnivcTestTaskTeam3/_apis/tfvc/items/$/Project2/test.txt?versionType=Changeset&version=17"

	//получаем результат работы функции
	addedRows, deletedRows, err := azure.ChangedRows(projectName, currentFileUrl, previousFileUrl)

	//проверки
	assert.NoError(t, err)
	assert.Equal(t, 4, addedRows)
	assert.Equal(t, 1, deletedRows)
}

func TestGetItemVersions(t *testing.T) {
	conf := NewConfig()
	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

	azure := NewAzure(conf)
	azure.Connect()
	azure.TfvcClientConnection()

	fmt.Println(azure.GetItemVersions("Project2"))
}
