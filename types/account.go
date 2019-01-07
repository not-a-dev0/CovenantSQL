/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"encoding/json"
	"sync"

	pi "github.com/CovenantSQL/CovenantSQL/blockproducer/interfaces"
	"github.com/CovenantSQL/CovenantSQL/proto"
)

//go:generate hsp
//hsp:ignore PermStat

// SQLChainRole defines roles of account in a SQLChain.
type SQLChainRole byte

const (
	// Miner defines the miner role as a SQLChain user.
	Miner SQLChainRole = iota
	// Customer defines the customer role as a SQLChain user.
	Customer
	// NumberOfRoles defines the SQLChain roles number.
	NumberOfRoles
)

// UserPermissionRole defines role of user permission including admin/write/read.
type UserPermissionRole int32

// UserPermission defines permissions of a SQLChain user.
type UserPermission struct {
	// User role to access database.
	Role UserPermissionRole
	// SQL pattern regulations for user queries
	// only a fully matched (case-sensitive) sql query is permitted to execute.
	Patterns []string

	// patterns map cache for matching
	cachedPatternMapOnce sync.Once
	cachedPatternMap     map[string]bool
}

const (
	// Void defines the initial permission.
	Void UserPermissionRole = iota
	// Admin defines the admin user permission.
	Admin
	// Write defines the writer user permission.
	Write
	// Read defines the reader user permission.
	Read
	// NumberOfUserPermission defines the user permission number.
	NumberOfUserPermission
)

// UnmarshalJSON implements the json.Unmarshler interface.
func (r *UserPermissionRole) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	r.FromString(s)
	return
}

// MarshalJSON implements the json.Marshaler interface.
func (r UserPermissionRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// String implements the fmt.Stringer interface.
func (r UserPermissionRole) String() string {
	switch r {
	case Admin:
		return "Admin"
	case Write:
		return "Write"
	case Read:
		return "Read"
	case Void:
		return "Void"
	default:
		return "Unknown"
	}
}

// FromString converts string to UserPermissionRole.
func (r *UserPermissionRole) FromString(perm string) {
	switch perm {
	case "Admin":
		*r = Admin
	case "Write":
		*r = Write
	case "Read":
		*r = Read
	case "Void":
		*r = Void
	default:
		*r = NumberOfUserPermission
	}
}

// UserPermissionFromRole construct a new user permission instance from primitive user permission role enum.
func UserPermissionFromRole(role UserPermissionRole) *UserPermission {
	return &UserPermission{
		Role: role,
	}
}

// HasReadPermission returns true if user owns read permission.
func (up *UserPermission) HasReadPermission() bool {
	if up == nil {
		return false
	}
	return up.Role >= Admin && up.Role < NumberOfUserPermission
}

// HasWritePermission returns true if user owns write permission.
func (up *UserPermission) HasWritePermission() bool {
	if up == nil {
		return false
	}
	return up.Role >= Admin && up.Role <= Write
}

// HasAdminPermission returns true if user owns admin permission.
func (up *UserPermission) HasAdminPermission() bool {
	if up == nil {
		return false
	}
	return up.Role == Admin
}

// IsValid returns whether the permission object is valid or not.
func (up *UserPermission) IsValid() bool {
	return up != nil && up.Role < NumberOfUserPermission && up.Role >= Admin
}

// HasDisallowedQueryPatterns returns whether the queries are permitted.
func (up *UserPermission) HasDisallowedQueryPatterns(queries []Query) (query string, status bool) {
	if up == nil {
		status = true
		return
	}
	if len(up.Patterns) == 0 {
		status = false
		return
	}

	up.cachedPatternMapOnce.Do(func() {
		up.cachedPatternMap = make(map[string]bool, len(up.Patterns))
		for _, p := range up.Patterns {
			up.cachedPatternMap[p] = true
		}
	})

	for _, q := range queries {
		if !up.cachedPatternMap[q.Pattern] {
			// not permitted
			query = q.Pattern
			status = true
			break
		}
	}

	return
}

// Status defines status of a SQLChain user/miner.
type Status int32

const (
	// UnknownStatus defines initial status.
	UnknownStatus Status = iota
	// Normal defines no bad thing happens.
	Normal
	// Reminder defines the user needs to increase advance payment.
	Reminder
	// Arrears defines the user is in arrears.
	Arrears
	// Arbitration defines the user/miner is in an arbitration.
	Arbitration
	// NumberOfStatus defines the number of status.
	NumberOfStatus
)

// EnableQuery indicates whether the account is permitted to query.
func (s *Status) EnableQuery() bool {
	return *s >= Normal && *s <= Reminder
}

// PermStat defines the permissions status structure.
type PermStat struct {
	Permission *UserPermission
	Status     Status
}

// SQLChainUser defines a SQLChain user.
type SQLChainUser struct {
	Address        proto.AccountAddress
	Permission     *UserPermission
	AdvancePayment uint64
	Arrears        uint64
	Deposit        uint64
	Status         Status
}

// UserArrears defines user's arrears.
type UserArrears struct {
	User    proto.AccountAddress
	Arrears uint64
}

// MinerInfo defines a miner.
type MinerInfo struct {
	Address        proto.AccountAddress
	NodeID         proto.NodeID
	Name           string
	PendingIncome  uint64
	ReceivedIncome uint64
	UserArrears    []*UserArrears
	Deposit        uint64
	Status         Status
	EncryptionKey  string
}

// SQLChainProfile defines a SQLChainProfile related to an account.
type SQLChainProfile struct {
	ID                proto.DatabaseID
	Address           proto.AccountAddress
	Period            uint64
	GasPrice          uint64
	LastUpdatedHeight uint32

	TokenType TokenType

	Owner proto.AccountAddress
	// first miner in the list is leader
	Miners []*MinerInfo

	Users []*SQLChainUser

	EncodedGenesis []byte

	Meta ResourceMeta // dumped from db creation tx
}

// ProviderProfile defines a provider list.
type ProviderProfile struct {
	Provider      proto.AccountAddress
	Space         uint64  // reserved storage space in bytes
	Memory        uint64  // reserved memory in bytes
	LoadAvgPerCPU float64 // max loadAvg15 per CPU
	TargetUser    []proto.AccountAddress
	Deposit       uint64 // default 10 Particle
	GasPrice      uint64
	TokenType     TokenType // default Particle
	NodeID        proto.NodeID
}

// Account store its balance, and other mate data.
type Account struct {
	Address      proto.AccountAddress
	TokenBalance [SupportTokenNumber]uint64
	Rating       float64
	NextNonce    pi.AccountNonce
}
