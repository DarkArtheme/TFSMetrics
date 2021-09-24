package azure

import (
	"io"
	"strings"
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

func (a *Azure) ChangedRows(currentFileUrl string) (int, int, error) {
	//1 GET FILES
	//get current version
	item, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl})
	if err != nil {
		return 0, 0, err
	}
	b1, err := io.ReadAll(item)
	if err != nil { //Read All File in array byte
		return 0, 0, err
	}
	//get previous version
	item1, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl,
		VersionDescriptor: &git.TfvcVersionDescriptor{VersionOption: &git.TfvcVersionOptionValues.Previous}})
	if err != nil {
		return a.getAddedRowsOneFile(&b1) //если нет прошлой версии считаем кол-во строк в текущем файле
	}
	b2, err := io.ReadAll(item1)
	if err != nil {
		return 0, 0, err
	}

	//считаем добаленные и удаленные строки
	addedRows, deletedRows := Diff(string(b2), string(b1))
	return addedRows, deletedRows, nil
}

func (a *Azure) getAddedRowsOneFile(fileInBytes *[]byte) (int, int, error) {
	transformString := string(*fileInBytes)                     //transform array byte in string
	arrTransformStrings := strings.Split(transformString, "\n") //split string
	return len(arrTransformStrings), 0, nil
}

// заглушка чтобы избавиться от ошибки нереализованного интерфейса
func (a *Azure) GetItemVersions(ChangesUrl string) (int, int) {
	changesets, _ := a.GetChangesets(ChangesUrl)
	if len(changesets) > 1 {
		return *(changesets)[0], *(changesets)[1]
	}

	return *(changesets)[0], 0
}

//diff
type Chunk struct {
	Added   []string
	Deleted []string
	Equal   []string
}

func (c *Chunk) empty() bool {
	return len(c.Added) == 0 && len(c.Deleted) == 0 && len(c.Equal) == 0
}

// Diff returns a string containing a line-by-line unified diff of the linewise
// changes required to make A into B.  Each line is prefixed with '+', '-', or
// ' ' to indicate if it should be added, removed, or is correct respectively.
func Diff(A, B string) (int, int) {
	aLines := strings.Split(A, "\n")
	bLines := strings.Split(B, "\n")
	return Render(DiffChunks(aLines, bLines))
}

// Render renders the slice of chunks into a representation that prefixes
// the lines with '+', '-', or ' ' depending on whether the line was added,
// removed, or equal (respectively).
func Render(chunks []Chunk) (int, int) {
	addedRows := 0
	deletedRows := 0

	for _, v := range chunks {
		if len(v.Added) == 1 {
			addedRows++
		} else if len(v.Deleted) == 1 {
			deletedRows++
		}
	}
	return addedRows, deletedRows
}

// DiffChunks uses an O(D(N+M)) shortest-edit-script algorithm
// to compute the edits required from A to B and returns the
// edit chunks.
func DiffChunks(a, b []string) []Chunk {
	// algorithm: http://www.xmailserver.org/diff2.pdf

	// We'll need these quantities a lot.
	alen, blen := len(a), len(b) // M, N

	// At most, it will require len(a) deletions and len(b) additions
	// to transform a into b.
	maxPath := alen + blen // MAX
	if maxPath == 0 {
		// degenerate case: two empty lists are the same
		return nil
	}

	// Store the endpoint of the path for diagonals.
	// We store only the a index, because the b index on any diagonal
	// (which we know during the loop below) is aidx-diag.
	// endpoint[maxPath] represents the 0 diagonal.
	//
	// Stated differently:
	// endpoint[d] contains the aidx of a furthest reaching path in diagonal d
	endpoint := make([]int, 2*maxPath+1) // V

	saved := make([][]int, 0, 8) // Vs
	save := func() {
		dup := make([]int, len(endpoint))
		copy(dup, endpoint)
		saved = append(saved, dup)
	}

	var editDistance int // D
dLoop:
	for editDistance = 0; editDistance <= maxPath; editDistance++ {
		// The 0 diag(onal) represents equality of a and b.  Each diagonal to
		// the left is numbered one lower, to the right is one higher, from
		// -alen to +blen.  Negative diagonals favor differences from a,
		// positive diagonals favor differences from b.  The edit distance to a
		// diagonal d cannot be shorter than d itself.
		//
		// The iterations of this loop cover either odds or evens, but not both,
		// If odd indices are inputs, even indices are outputs and vice versa.
		for diag := -editDistance; diag <= editDistance; diag += 2 { // k
			var aidx int // x
			switch {
			case diag == -editDistance:
				// This is a new diagonal; copy from previous iter
				aidx = endpoint[maxPath-editDistance+1] + 0
			case diag == editDistance:
				// This is a new diagonal; copy from previous iter
				aidx = endpoint[maxPath+editDistance-1] + 1
			case endpoint[maxPath+diag+1] > endpoint[maxPath+diag-1]:
				// diagonal d+1 was farther along, so use that
				aidx = endpoint[maxPath+diag+1] + 0
			default:
				// diagonal d-1 was farther (or the same), so use that
				aidx = endpoint[maxPath+diag-1] + 1
			}
			// On diagonal d, we can compute bidx from aidx.
			bidx := aidx - diag // y
			// See how far we can go on this diagonal before we find a difference.
			for aidx < alen && bidx < blen && a[aidx] == b[bidx] {
				aidx++
				bidx++
			}
			// Store the end of the current edit chain.
			endpoint[maxPath+diag] = aidx
			// If we've found the end of both inputs, we're done!
			if aidx >= alen && bidx >= blen {
				save() // save the final path
				break dLoop
			}
		}
		save() // save the current path
	}
	if editDistance == 0 {
		return nil
	}
	chunks := make([]Chunk, editDistance+1)

	x, y := alen, blen
	for d := editDistance; d > 0; d-- {
		endpoint := saved[d]
		diag := x - y
		insert := diag == -d || (diag != d && endpoint[maxPath+diag-1] < endpoint[maxPath+diag+1])

		x1 := endpoint[maxPath+diag]
		var x0, xM, kk int
		if insert {
			kk = diag + 1
			x0 = endpoint[maxPath+kk]
			xM = x0
		} else {
			kk = diag - 1
			x0 = endpoint[maxPath+kk]
			xM = x0 + 1
		}
		y0 := x0 - kk

		var c Chunk
		if insert {
			c.Added = b[y0:][:1]
		} else {
			c.Deleted = a[x0:][:1]
		}
		if xM < x1 {
			c.Equal = a[xM:][:x1-xM]
		}

		x, y = x0, y0
		chunks[d] = c
	}
	if x > 0 {
		chunks[0].Equal = a[:x]
	}
	if chunks[0].empty() {
		chunks = chunks[1:]
	}
	if len(chunks) == 0 {
		return nil
	}
	return chunks
}
