package mzk

const (
	EventCreate      int = 1 << iota // 1
	EventDelete                      // 10
	EventDataChange                  // 100
	EventChildChange                 // 1000
)
