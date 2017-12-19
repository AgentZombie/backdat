package backdat

import "time"

type FP interface {
	AddFingerprint(string) error
	HaveFingerprint(string) (bool, error)
	Close() error
}

type FPStore interface {
	ListFingerprintSets() ([]time.Time, error)
	OpenFingerprintSet(time.Time) (FP, error)
}
