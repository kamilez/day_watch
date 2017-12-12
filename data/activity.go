package data

type Activity struct {
	Start string `sql:"TEXT NOT NULL"`
	Stop  string `sql:"TEXT NOT NULL"`
	Date  string `sql:"TEXT NOT NULL"`
	Type  string `sql:"TEXT NOT NULL"`
}
