package core

type FunctionAppRepository interface {
	Save(app *FunctionApp) error
	GetByID(id string) (*FunctionApp, error)
	GetAll() ([]*FunctionApp, error)
	Delete(id string) error
}

type FunctionCompositionRepository interface {
	Save(comp *FunctionComposition) error
	GetByID(id string) (*FunctionComposition, error)
	Delete(id string) error
}
