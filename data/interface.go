/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package data

// Service interface ensures that all services can be started and stopped in a loop
type Service interface {
	Start() error
	Stop() error
	Name() string
}