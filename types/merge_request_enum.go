package types

import "fmt"

const (
	MergeRequestStateUnknown MergeRequestState = iota
	MergeRequestStateOpen
	MergeRequestStateMerged
	MergeRequestStateClosed
)

const _MergeRequestState_name = "UnknownOpenMergedClosed"

var _MergeRequestStateMap = map[MergeRequestState]string{
	MergeRequestStateUnknown: _MergeRequestState_name[0:7],
	MergeRequestStateOpen:    _MergeRequestState_name[7:11],
	MergeRequestStateMerged:  _MergeRequestState_name[11:17],
	MergeRequestStateClosed:  _MergeRequestState_name[17:23],
}

var _MergeRequestStateValue = map[string]MergeRequestState{
	_MergeRequestState_name[0:7]:   MergeRequestStateUnknown,
	_MergeRequestState_name[7:11]:  MergeRequestStateOpen,
	_MergeRequestState_name[11:17]: MergeRequestStateMerged,
	_MergeRequestState_name[17:23]: MergeRequestStateClosed,
}

func (i MergeRequestState) String() string {
	if str, ok := _MergeRequestStateMap[i]; ok {
		return str
	}
	return fmt.Sprintf("MergeRequestState(%d)", i)
}

func MergeRequestStateFromString(s string) (MergeRequestState, error) {
	if v, ok := _MergeRequestStateValue[s]; ok {
		return v, nil
	}
	return MergeRequestStateUnknown, fmt.Errorf("invalid MergeRequestState: %q", s)
}
