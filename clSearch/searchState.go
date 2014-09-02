package clSearch

import "sync"

const (
	NoSearch = iota
	SearchInitialized = iota
	SearchInProgress = iota
	SearchDisposing = iota
	SearchCompleted = iota
)
type searchState struct {
	stateMutex sync.Mutex
	stateValue int
}

func (st searchState) getSearchState() int {
	st.stateMutex.Lock()
	defer st.stateMutex.Unlock()
	return st.stateValue
}

func (st searchState) setSearchState (state int) {
	st.stateMutex.Lock()
	defer st.stateMutex.Unlock()
	if state < NoSearch || state > SearchCompleted {
		panic("Invalid transition!")
	}
}


