package balancer

type Balancer interface {
	Next() *Backend
}
