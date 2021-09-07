package fielparser

type ParserInterface interface {
	// принимает ссылки на разные версии файлов возвращает Добавленные и Удаленные строки
	ChangedRows(currentFielUrel string, PreviusFileUrl string) (int, int)
}
