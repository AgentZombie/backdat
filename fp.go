package backdat

import "time"

type FP interface {
	AddFingerprint(string) error
	Close() error
	HaveFingerprint(string) (bool, error)
}

type FPStore interface {
	ListFingerprintSets() ([]time.Time, error)
	NewFingerprintSet(time.Time) (FP, error)
	OpenFingerprintSet(time.Time) (FP, error)
}

type NilFP struct{}

func (f NilFP) AddFingerprint(unused string) error {
	return nil
}

func (f NilFP) Close() error {
	return nil
}

func (f NilFP) HaveFingerprint(string) (bool, error) {
	return false, nil
}
