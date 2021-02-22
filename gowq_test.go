package gowq

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWQ(t *testing.T) {
	wq := NewWQ(5)

	require.Equal(t, wq.nWorkers, 5, "Unexpected number of workers")
}

func TestGetErrors(t *testing.T) {
	wq := NewWQ(0)

	errorsList := wq.GetErrors(false)
	require.Equal(t, 0, len(errorsList), "Unexpected returned number of errors")
	require.Equal(t, 0, len(wq.errorsList), "Unexpected internal number of errors")

	wq.errorsList = append(wq.errorsList, fmt.Errorf("new error"))

	errorsList = wq.GetErrors(false)
	require.Equal(t, 1, len(errorsList), "Unexpected returned number of errors")
	require.Equal(t, 1, len(wq.errorsList), "Unexpected internal number of errors")

	errorsList = wq.GetErrors(true)
	require.Equal(t, 1, len(errorsList), "Unexpected returned number of errors")
	require.Equal(t, 0, len(wq.errorsList), "Unexpected internal number of errors")
}
