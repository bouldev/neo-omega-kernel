package blocks

import (
	"fmt"
	"sync"
)

const UNKNOWM_RUNTIME = uint32(0xFFFFFFFF)

type ToNEMCBaseNames struct {
	legacyValuesMapping []uint32
	statesWithRtid      []struct {
		states *PropsForSearch
		rtid   uint32
	}
	StatesWithRtidQuickMatch map[string]uint32
	mu                       sync.RWMutex
}

func (baseNameGroup *ToNEMCBaseNames) addAnchorByLegacyValue(legacyValue int16, nemcRTID uint32) (exist bool, conflictErr error) {
	if int(legacyValue+1) <= len(baseNameGroup.legacyValuesMapping) {
		if baseNameGroup.legacyValuesMapping[legacyValue] == UNKNOWM_RUNTIME {
			baseNameGroup.legacyValuesMapping[legacyValue] = nemcRTID
			return false, nil
		} else if baseNameGroup.legacyValuesMapping[legacyValue] != nemcRTID {
			return true, fmt.Errorf("conflict runtime id ")
		} else {
			return true, nil
		}
	}
	baseNameGroup.mu.Lock()
	defer baseNameGroup.mu.Unlock()
	for int(legacyValue+1) > len(baseNameGroup.legacyValuesMapping) {
		baseNameGroup.legacyValuesMapping = append(baseNameGroup.legacyValuesMapping, UNKNOWM_RUNTIME)
	}
	baseNameGroup.legacyValuesMapping[legacyValue] = nemcRTID
	return false, nil
}

func (baseNameGroup *ToNEMCBaseNames) preciseMatchByLegacyValue(legacyValue int16) (rtid uint32, found bool) {
	if int(legacyValue+1) <= len(baseNameGroup.legacyValuesMapping) {
		if rtid = baseNameGroup.legacyValuesMapping[legacyValue]; rtid == UNKNOWM_RUNTIME {
			return uint32(AIR_RUNTIMEID), false
		} else {
			return rtid, true
		}
	} else {
		return uint32(AIR_RUNTIMEID), false
	}
}

func (baseNameGroup *ToNEMCBaseNames) fuzzySearchByLegacyValue(legacyValue int16) (rtid uint32, found bool) {
	if int(legacyValue+1) <= len(baseNameGroup.legacyValuesMapping) {
		if rtid = baseNameGroup.legacyValuesMapping[legacyValue]; rtid != UNKNOWM_RUNTIME {
			return rtid, true
		}
	}
	if int(legacyValue+1) <= len(baseNameGroup.statesWithRtid) {
		return baseNameGroup.statesWithRtid[legacyValue].rtid, true
	}
	return baseNameGroup.statesWithRtid[0].rtid, true
}

func (baseNameGroup *ToNEMCBaseNames) addAnchorByState(states *PropsForSearch, runtimeID uint32, overwrite bool) (exist bool, conflictErr error) {
	quickMatchStr := "{}"
	if states != nil {
		quickMatchStr = states.InPreciseSNBT()
	}
	baseNameGroup.mu.RLock()
	if currentRuntimeID, found := baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr]; found {
		if currentRuntimeID == runtimeID {
			baseNameGroup.mu.RUnlock()
			return true, nil
		} else if !overwrite {
			baseNameGroup.mu.RUnlock()
			return true, fmt.Errorf("conflict runtime id ")
		}
	}
	baseNameGroup.mu.RUnlock()
	baseNameGroup.mu.Lock()
	defer baseNameGroup.mu.Unlock()
	baseNameGroup.statesWithRtid = append(baseNameGroup.statesWithRtid, struct {
		states *PropsForSearch
		rtid   uint32
	}{states: states, rtid: runtimeID})
	baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr] = runtimeID
	return false, nil
}

func (baseNameGroup *ToNEMCBaseNames) preciseMatchByState(states *PropsForSearch) (rtid uint32, found bool) {
	quickMatchStr := states.InPreciseSNBT()
	baseNameGroup.mu.RLock()
	defer baseNameGroup.mu.RUnlock()
	rtid, found = baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr]
	return rtid, found
}

func (baseNameGroup *ToNEMCBaseNames) fuzzySearchByState(states *PropsForSearch) (rtid uint32, score ComparedOutput, matchAny bool) {
	quickMatchStr := states.InPreciseSNBT()
	baseNameGroup.mu.RLock()
	defer baseNameGroup.mu.RUnlock()
	rtid, found := baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr]
	if found {
		return rtid, ComparedOutput{Same: uint8(len(states.props))}, true
	}
	rtid = uint32(AIR_RUNTIMEID)
	matchAny = false
	for _, anchor := range baseNameGroup.statesWithRtid {
		matchAny = true
		newScore := anchor.states.Compare(states)
		if newScore.Same > score.Same || (newScore.Same == score.Same && ((newScore.Different + newScore.Redundant + newScore.Missing) < (score.Different + score.Redundant + score.Missing))) {
			score = newScore
			rtid = anchor.rtid
		}
	}
	return rtid, score, matchAny
}

type ToNEMCConverter struct {
	BaseNames map[string]*ToNEMCBaseNames
	mu        sync.RWMutex
}

func (c *ToNEMCConverter) ensureBaseNameGroup(name string) *ToNEMCBaseNames {
	c.mu.RLock()
	if to, found := c.BaseNames[name]; found {
		c.mu.RUnlock()
		return to
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	to := &ToNEMCBaseNames{
		legacyValuesMapping: make([]uint32, 0),
		statesWithRtid: make([]struct {
			states *PropsForSearch
			rtid   uint32
		}, 0),
		StatesWithRtidQuickMatch: make(map[string]uint32),
		mu:                       sync.RWMutex{},
	}
	c.BaseNames[name] = to
	return to
}

func (c *ToNEMCConverter) getBaseNameGroup(name string) (baseGroup *ToNEMCBaseNames, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	group, found := c.BaseNames[name]
	return group, found
}

func (c *ToNEMCConverter) AddAnchorByLegacyValue(name BaseWithNameSpace, legacyValue int16, nemcRTID uint32) (exist bool, conflictErr error) {
	baseNameGroup := c.ensureBaseNameGroup(name.BaseName())
	return baseNameGroup.addAnchorByLegacyValue(legacyValue, nemcRTID)
}

func (c *ToNEMCConverter) PreciseMatchByLegacyValue(name BaseWithNameSpace, legacyValue int16) (rtid uint32, found bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(NEMC_AIR_RUNTIMEID), false
	}
	return baseGroup.preciseMatchByLegacyValue(legacyValue)
}

func (c *ToNEMCConverter) TryBestSearchByLegacyValue(name BaseWithNameSpace, legacyValue int16) (rtid uint32, found bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(NEMC_AIR_RUNTIMEID), false
	}
	return baseGroup.fuzzySearchByLegacyValue(legacyValue)
}

func (c *ToNEMCConverter) AddAnchorByState(name BaseWithNameSpace, states *PropsForSearch, runtimeID uint32, overwrite bool) (exist bool, conflictErr error) {
	baseNameGroup := c.ensureBaseNameGroup(name.BaseName())
	return baseNameGroup.addAnchorByState(states, runtimeID, overwrite)
}

func (c *ToNEMCConverter) PreciseMatchByState(name BaseWithNameSpace, states *PropsForSearch) (rtid uint32, found bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(NEMC_AIR_RUNTIMEID), false
	}
	return baseGroup.preciseMatchByState(states)
}

func (c *ToNEMCConverter) TryBestSearchByState(name BaseWithNameSpace, states *PropsForSearch) (rtid uint32, score ComparedOutput, matchAny bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(NEMC_AIR_RUNTIMEID), ComparedOutput{}, false
	}
	return baseGroup.fuzzySearchByState(states)
}

var DefaultAnyToNemcConvertor = &ToNEMCConverter{
	BaseNames: map[string]*ToNEMCBaseNames{},
	mu:        sync.RWMutex{},
}
