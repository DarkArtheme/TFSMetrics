package azure

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
)

type AzureInterface interface {
	Azure() *Azure
	Connect()                           // Подключение к Azure DevOps
	TfvcClientConnection() error        // для Repository.Open()
	ListOfProjects() ([]*string, error) // Получаем список проектов

	GetChangesets(nameOfProject string) ([]*int, error)                          // Получает все id ченджсетов проекта
	GetChangesetChanges(id *int, project string) (*ChangeSet, error)             // получает все изминения для конкретного changeSet
	GetItemVersions(ChangesUrl string) (int, int)                                // Находит искомую и предыдущую версию файла, возвращает их юрл'ы
	ChangedRows(currentFileUrl string, PreviousFileUrl string) (int, int, error) // Принимает ссылки на разные версии файлов возвращает Добавленные и Удаленные строки
}

type ChangeSet struct {
	ProjectName string
	Id          int
	Author      string
	Email       string
	AddedRows   int
	DeletedRows int
	Date        time.Time
	Message     string
	Hash        string
}

type Azure struct {
	Config     *Config
	Connection *azuredevops.Connection
	TfvcClient tfvc.Client
}

func NewAzure(config *Config) *Azure {
	return &Azure{
		Config: config,
	}
}

func (a *Azure) Azure() *Azure {
	return a
}

func (a *Azure) Connect() {
	organizationUrl := a.Config.OrganizationUrl
	personalAccessToken := a.Config.Token

	connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)
	a.Connection = connection
}

func (a *Azure) TfvcClientConnection() error {
	tfvcClient, err := tfvc.NewClient(a.Config.Context, a.Connection)
	if err != nil {
		return err
	}
	a.TfvcClient = tfvcClient
	return nil
}

func (a *Azure) ListOfProjects() ([]*string, error) {
	coreClient, err := core.NewClient(a.Config.Context, a.Connection)
	if err != nil {
		return nil, err
	}

	resp, err := coreClient.GetProjects(a.Config.Context, core.GetProjectsArgs{})
	if err != nil {
		return nil, err
	}
	projectNames := []*string{}
	for _, project := range resp.Value {
		projectNames = append(projectNames, project.Name)
	}
	return projectNames, nil
}

func (a *Azure) GetChangesets(nameOfProject string) ([]*int, error) {
	changeSets, err := a.TfvcClient.GetChangesets(a.Config.Context, tfvc.GetChangesetsArgs{Project: &nameOfProject})
	if err != nil {
		return nil, err
	}
	changeSetIDs := []*int{}
	for _, v := range *changeSets {
		changeSetIDs = append(changeSetIDs, v.ChangesetId)
	}
	return changeSetIDs, nil
}

func (a *Azure) GetChangesetChanges(id *int, project string) (*ChangeSet, error) {
	changes, err := a.TfvcClient.GetChangeset(a.Config.Context, tfvc.GetChangesetArgs{Id: id, Project: &project})
	if err != nil {
		return &ChangeSet{}, err
	}
	messg := ""
	if changes.Comment != nil {
		messg = *changes.Comment
	}

	//changesHash, err := a.TfvcClient.GetChangesetChanges(a.Config.Context, tfvc.GetChangesetChangesArgs{Id: id})
	if err != nil {
		return nil, err
	}
	//fmt.Println(changesHash.Value[0].Item.(map[string]interface{}))

	commit := &ChangeSet{
		ProjectName: project,
		Id:          *id,
		Author:      *changes.Author.DisplayName,
		Email:       *changes.Author.UniqueName,
		Date:        changes.CreatedDate.Time,
		Message:     messg,
	}
	//fmt.Println(changesHash.Value[0].Item.(map[string]interface{})["version"])
	return commit, nil
}

func (a *Azure) ChangedRows(nameProject string, currentFileUrl string, PreviousFileUrl string) (int, int, error) {
	//1 GET FILES
	_, previousVersion := a.GetItemVersions(nameProject) //получить версии изменений
	if previousVersion == 0 {                            //посчитать изменения только в текущей версии
		return a.getAddedRowsOneFile(currentFileUrl)
	}

	item, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl})
	if err != nil {
		return 0, 0, err
	}

	b1, err := io.ReadAll(item)
	if err != nil { //Read All File in array byte
		return 0, 0, err
	}

	ver := fmt.Sprint(previousVersion) // заменить на код для разных версий
	item1, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl,
		VersionDescriptor: &git.TfvcVersionDescriptor{Version: &ver}})
	if err != nil {
		return 0, 0, err
	}
	b2, err := io.ReadAll(item1)
	if err != nil {
		return 0, 0, err
	}

	//2 конвертируем полученные данные в map
	wg := sync.WaitGroup{}
	wg.Add(2)
	var currentStringsMap, copyCurrentStringsMap, previousStringsMap, copyPreviousStringsMap map[int]string
	go transformArrByteToMap(&b1, &wg, &currentStringsMap, &copyCurrentStringsMap)
	go transformArrByteToMap(&b2, &wg, &previousStringsMap, &copyPreviousStringsMap)
	wg.Wait()

	//3 счатаем добавленные строки
	//counters
	addedRows := 0
	deletedRows := 0

	wg.Add(2)
	go countAddedRows(copyCurrentStringsMap, copyPreviousStringsMap, &addedRows, &wg) //count added strings
	go countDeleteRows(currentStringsMap, previousStringsMap, &deletedRows, &wg)      //count delete strings
	wg.Wait()

	return addedRows, deletedRows, err
}

func (a *Azure) getAddedRowsOneFile(currentFileUrl string) (int, int, error) {
	item, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl})
	if err != nil {
		return 0, 0, err
	}

	b1, err := io.ReadAll(item)
	if err != nil { //Read All File in array byte
		return 0, 0, err
	}

	transformString := string(b1)                               //transform array byte in string
	arrTransformStrings := strings.Split(transformString, "\n") //split string
	return len(arrTransformStrings), 0, nil
}

func transformArrByteToMap(arr *[]byte, wg *sync.WaitGroup, transformMap, transformMapCopy *map[int]string) {
	defer wg.Done()
	transformString := string(*arr)                             //transform array byte in string
	arrTransformStrings := strings.Split(transformString, "\n") //split string
	//create maps
	*transformMap = make(map[int]string, len(arrTransformStrings))
	*transformMapCopy = make(map[int]string, len(arrTransformStrings))
	//add strings in map
	for k, v := range arrTransformStrings {
		(*transformMap)[k] = v
		(*transformMapCopy)[k] = v
	}
}

func countAddedRows(currentStringsMap, previousStringsMap map[int]string, addedRows *int, wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0 //for chek end currentStrings
	for _, v1 := range currentStringsMap {
		i = 0
		for k, v2 := range previousStringsMap {
			if v1 == v2 {
				delete(previousStringsMap, k)
				i--
				break
			}
			i++
		}
		if i == len(previousStringsMap) {
			(*addedRows)++
		}
	}
}

func countDeleteRows(currentStringsMap, previousStringsMap map[int]string, deletedRows *int, wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0 //for chek end currentStrings
	for _, v1 := range previousStringsMap {
		i = 0
		for k, v2 := range currentStringsMap {
			if v1 == v2 {
				delete(currentStringsMap, k)
				i--
				break
			}
			i++
		}
		if i == len(currentStringsMap) {
			(*deletedRows)++
		}
	}
}

// заглушка чтобы избавиться от ошибки нереализованного интерфейса
func (a *Azure) GetItemVersions(ChangesUrl string) (int, int) {
	changesets, _ := a.GetChangesets(ChangesUrl)
	if len(changesets) > 1 {
		return *(changesets)[0], *(changesets)[1]
	}

	return *(changesets)[0], 0
}
