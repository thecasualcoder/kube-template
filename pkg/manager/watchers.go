package manager

type watchers struct {
	data *Store
}

func newWatchers() *watchers {
	return &watchers{NewStore()}
}

func (w *watchers) exists(key string) bool {
	_, present := w.data.Get(key)
	return present
}

func (w *watchers) add(key string) {
	w.data.Set(key, struct{}{})
}
