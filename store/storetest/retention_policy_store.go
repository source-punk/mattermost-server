// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"sort"
	"strconv"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/stretchr/testify/require"
)

func TestRetentionPolicyStore(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("Save", func(t *testing.T) { testRetentionPolicyStoreSave(t, ss, s) })
	t.Run("Patch", func(t *testing.T) { testRetentionPolicyStorePatch(t, ss, s) })
	t.Run("Get", func(t *testing.T) { testRetentionPolicyStoreGet(t, ss, s) })
	t.Run("GetCount", func(t *testing.T) { testRetentionPolicyStoreGetCount(t, ss, s) })
	t.Run("Delete", func(t *testing.T) { testRetentionPolicyStoreDelete(t, ss, s) })
	t.Run("GetChannels", func(t *testing.T) { testRetentionPolicyStoreGetChannels(t, ss, s) })
	t.Run("AddChannels", func(t *testing.T) { testRetentionPolicyStoreAddChannels(t, ss, s) })
	t.Run("RemoveChannels", func(t *testing.T) { testRetentionPolicyStoreRemoveChannels(t, ss, s) })
	t.Run("GetTeams", func(t *testing.T) { testRetentionPolicyStoreGetTeams(t, ss, s) })
	t.Run("AddTeams", func(t *testing.T) { testRetentionPolicyStoreAddTeams(t, ss, s) })
	t.Run("RemoveTeams", func(t *testing.T) { testRetentionPolicyStoreRemoveTeams(t, ss, s) })
}

func getRetentionPolicyWithTeamAndChannelIds(t *testing.T, ss store.Store, policyID string) *model.RetentionPolicyWithTeamAndChannelIDs {
	policyWithCounts, err := ss.RetentionPolicy().Get(policyID)
	require.NoError(t, err)
	policyWithIds := model.RetentionPolicyWithTeamAndChannelIDs{
		RetentionPolicy: model.RetentionPolicy{
			ID:           policyID,
			DisplayName:  policyWithCounts.DisplayName,
			PostDuration: policyWithCounts.PostDuration,
		},
		ChannelIDs: make([]string, int(policyWithCounts.ChannelCount)),
		TeamIDs:    make([]string, int(policyWithCounts.TeamCount)),
	}
	channels, err := ss.RetentionPolicy().GetChannels(policyID, 0, 1000)
	require.NoError(t, err)
	for i, channel := range channels {
		policyWithIds.ChannelIDs[i] = channel.Id
	}
	teams, err := ss.RetentionPolicy().GetTeams(policyID, 0, 1000)
	require.NoError(t, err)
	for i, team := range teams {
		policyWithIds.TeamIDs[i] = team.Id
	}
	return &policyWithIds
}

func CheckRetentionPolicyWithTeamAndChannelIdsAreEqual(t *testing.T, p1, p2 *model.RetentionPolicyWithTeamAndChannelIDs) {
	require.Equal(t, p1.ID, p2.ID)
	require.Equal(t, p1.DisplayName, p2.DisplayName)
	require.Equal(t, p1.PostDuration, p2.PostDuration)
	require.Equal(t, len(p1.ChannelIDs), len(p2.ChannelIDs))
	if p1.ChannelIDs == nil || p2.ChannelIDs == nil {
		require.Equal(t, p1.ChannelIDs, p2.ChannelIDs)
	} else {
		sort.Strings(p1.ChannelIDs)
		sort.Strings(p2.ChannelIDs)
	}
	for i := range p1.ChannelIDs {
		require.Equal(t, p1.ChannelIDs[i], p2.ChannelIDs[i])
	}
	if p1.TeamIDs == nil || p2.TeamIDs == nil {
		require.Equal(t, p1.TeamIDs, p2.TeamIDs)
	} else {
		sort.Strings(p1.TeamIDs)
		sort.Strings(p2.TeamIDs)
	}
	require.Equal(t, len(p1.TeamIDs), len(p2.TeamIDs))
	for i := range p1.TeamIDs {
		require.Equal(t, p1.TeamIDs[i], p2.TeamIDs[i])
	}
}

func CheckRetentionPolicyWithTeamAndChannelCountsAreEqual(t *testing.T, p1, p2 *model.RetentionPolicyWithTeamAndChannelCounts) {
	require.Equal(t, p1.ID, p2.ID)
	require.Equal(t, p1.DisplayName, p2.DisplayName)
	require.Equal(t, p1.PostDuration, p2.PostDuration)
	require.Equal(t, p1.ChannelCount, p2.ChannelCount)
	require.Equal(t, p1.TeamCount, p2.TeamCount)
}

func checkRetentionPolicyLikeThisExists(t *testing.T, ss store.Store, expected *model.RetentionPolicyWithTeamAndChannelIDs) {
	retrieved := getRetentionPolicyWithTeamAndChannelIds(t, ss, expected.ID)
	CheckRetentionPolicyWithTeamAndChannelIdsAreEqual(t, expected, retrieved)
}

func copyRetentionPolicyWithTeamAndChannelIds(policy *model.RetentionPolicyWithTeamAndChannelIDs) *model.RetentionPolicyWithTeamAndChannelIDs {
	copy := &model.RetentionPolicyWithTeamAndChannelIDs{
		RetentionPolicy: policy.RetentionPolicy,
		ChannelIDs:      make([]string, len(policy.ChannelIDs)),
		TeamIDs:         make([]string, len(policy.TeamIDs)),
	}
	for i, channelID := range policy.ChannelIDs {
		copy.ChannelIDs[i] = channelID
	}
	for i, teamID := range policy.TeamIDs {
		copy.TeamIDs[i] = teamID
	}
	return copy
}

func createChannelsForRetentionPolicy(t *testing.T, ss store.Store, teamId string, numChannels int) (channelIDs []string) {
	channelIDs = make([]string, numChannels)
	for i := range channelIDs {
		name := "channel" + model.NewId()
		channel := &model.Channel{
			TeamId:      teamId,
			DisplayName: "Channel " + name,
			Name:        name,
			Type:        model.CHANNEL_OPEN,
		}
		channel, err := ss.Channel().Save(channel, -1)
		require.Nil(t, err)
		channelIDs[i] = channel.Id
	}
	return
}

func createTeamsForRetentionPolicy(t *testing.T, ss store.Store, numTeams int) (teamIDs []string) {
	teamIDs = make([]string, numTeams)
	for i := range teamIDs {
		name := "team" + model.NewId()
		team := &model.Team{
			DisplayName: "Team " + name,
			Name:        name,
			Type:        model.TEAM_OPEN,
		}
		team, err := ss.Team().Save(team)
		require.Nil(t, err)
		teamIDs[i] = team.Id
	}
	return
}

func createTeamsAndChannelsForRetentionPolicy(t *testing.T, ss store.Store) (teamIDs, channelIDs []string) {
	teamIDs = createTeamsForRetentionPolicy(t, ss, 2)
	channels1 := createChannelsForRetentionPolicy(t, ss, teamIDs[0], 1)
	channels2 := createChannelsForRetentionPolicy(t, ss, teamIDs[1], 2)
	channelIDs = append(channels1, channels2...)
	return
}

func cleanupRetentionPolicyTest(s SqlStore) {
	// Manually clear tables until testlib can handle cleanups
	if _, err := s.GetMaster().Exec("DELETE FROM Channels"); err != nil {
		panic(err)
	}
	if _, err := s.GetMaster().Exec("DELETE FROM Teams"); err != nil {
		panic(err)
	}
	if _, err := s.GetMaster().Exec("DELETE FROM RetentionPolicies"); err != nil {
		panic(err)
	}
	if _, err := s.GetMaster().Exec("DELETE FROM RetentionPoliciesChannels"); err != nil {
		panic(err)
	}
	if _, err := s.GetMaster().Exec("DELETE FROM RetentionPoliciesTeams"); err != nil {
		panic(err)
	}
}

func createRetentionPolicyWithTeamAndChannelIds(displayName string, teamIDs, channelIDs []string) *model.RetentionPolicyWithTeamAndChannelIDs {
	return &model.RetentionPolicyWithTeamAndChannelIDs{
		RetentionPolicy: model.RetentionPolicy{
			DisplayName:  displayName,
			PostDuration: 30,
		},
		TeamIDs:    teamIDs,
		ChannelIDs: channelIDs,
	}
}

// saveRetentionPolicyWithTeamAndChannelIds creates a model.RetentionPolicyWithTeamAndChannelIds struct using
// the display name, team IDs, and channel IDs. The new policy ID will be assigned to the struct and returned.
// The team IDs and channel IDs are kept the same.
func saveRetentionPolicyWithTeamAndChannelIds(t *testing.T, ss store.Store, displayName string, teamIDs, channelIDs []string) *model.RetentionPolicyWithTeamAndChannelIDs {
	proposal := createRetentionPolicyWithTeamAndChannelIds(displayName, teamIDs, channelIDs)
	policyWithCounts, err := ss.RetentionPolicy().Save(proposal)
	require.Nil(t, err)
	proposal.ID = policyWithCounts.ID
	return proposal
}

func restoreRetentionPolicy(t *testing.T, ss store.Store, policy *model.RetentionPolicyWithTeamAndChannelIDs) {
	_, err := ss.RetentionPolicy().Patch(policy)
	require.Nil(t, err)
	checkRetentionPolicyLikeThisExists(t, ss, policy)
}

func testRetentionPolicyStoreSave(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("teams and channels are nil", func(t *testing.T) {
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Proposal 1", nil, nil)
		expected := createRetentionPolicyWithTeamAndChannelIds(policy.DisplayName, []string{}, []string{})
		checkRetentionPolicyLikeThisExists(t, ss, expected)
	})
	t.Run("teams and channels are empty", func(t *testing.T) {
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 2", []string{}, []string{})
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	t.Run("some teams and channels are specified", func(t *testing.T) {
		teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Proposal 3", teamIDs, channelIDs)
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	t.Run("team specified does not exist", func(t *testing.T) {
		policy := createRetentionPolicyWithTeamAndChannelIds("Policy 4", []string{"no_such_team"}, []string{})
		_, err := ss.RetentionPolicy().Save(policy)
		require.Error(t, err)
	})
	t.Run("channel specified does not exist", func(t *testing.T) {
		policy := createRetentionPolicyWithTeamAndChannelIds("Policy 5", []string{}, []string{"no_such_channel"})
		_, err := ss.RetentionPolicy().Save(policy)
		require.Error(t, err)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStorePatch(t *testing.T, ss store.Store, s SqlStore) {
	teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
	policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Proposal 1", teamIDs, channelIDs)
	t.Run("modify DisplayName", func(t *testing.T) {
		patch := &model.RetentionPolicyWithTeamAndChannelIDs{
			RetentionPolicy: model.RetentionPolicy{
				ID:          policy.ID,
				DisplayName: "something new",
			},
		}
		_, err := ss.RetentionPolicy().Patch(patch)
		require.Nil(t, err)
		expected := copyRetentionPolicyWithTeamAndChannelIds(policy)
		expected.DisplayName = patch.DisplayName
		checkRetentionPolicyLikeThisExists(t, ss, expected)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("modify PostDuration", func(t *testing.T) {
		patch := &model.RetentionPolicyWithTeamAndChannelIDs{
			RetentionPolicy: model.RetentionPolicy{
				ID:           policy.ID,
				PostDuration: 10000,
			},
		}
		_, err := ss.RetentionPolicy().Patch(patch)
		require.Nil(t, err)
		expected := copyRetentionPolicyWithTeamAndChannelIds(policy)
		expected.PostDuration = patch.PostDuration
		checkRetentionPolicyLikeThisExists(t, ss, expected)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("clear TeamIds", func(t *testing.T) {
		patch := &model.RetentionPolicyWithTeamAndChannelIDs{
			RetentionPolicy: model.RetentionPolicy{
				ID: policy.ID,
			},
			TeamIDs: make([]string, 0),
		}
		_, err := ss.RetentionPolicy().Patch(patch)
		require.Nil(t, err)
		expected := copyRetentionPolicyWithTeamAndChannelIds(policy)
		expected.TeamIDs = make([]string, 0)
		checkRetentionPolicyLikeThisExists(t, ss, expected)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("add team which does not exist", func(t *testing.T) {
		patch := &model.RetentionPolicyWithTeamAndChannelIDs{
			RetentionPolicy: model.RetentionPolicy{
				ID: policy.ID,
			},
			TeamIDs: []string{"no_such_team"},
		}
		_, err := ss.RetentionPolicy().Patch(patch)
		require.NotNil(t, err)
	})
	t.Run("clear ChannelIds", func(t *testing.T) {
		patch := &model.RetentionPolicyWithTeamAndChannelIDs{
			RetentionPolicy: model.RetentionPolicy{
				ID: policy.ID,
			},
			ChannelIDs: make([]string, 0),
		}
		_, err := ss.RetentionPolicy().Patch(patch)
		require.Nil(t, err)
		expected := copyRetentionPolicyWithTeamAndChannelIds(policy)
		expected.ChannelIDs = make([]string, 0)
		checkRetentionPolicyLikeThisExists(t, ss, expected)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("add channel which does not exist", func(t *testing.T) {
		patch := &model.RetentionPolicyWithTeamAndChannelIDs{
			RetentionPolicy: model.RetentionPolicy{
				ID: policy.ID,
			},
			ChannelIDs: []string{"no_such_channel"},
		}
		_, err := ss.RetentionPolicy().Patch(patch)
		require.NotNil(t, err)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreGet(t *testing.T, ss store.Store, s SqlStore) {
	// create multiple policies
	policiesWithCounts := make([]*model.RetentionPolicyWithTeamAndChannelCounts, 0)
	for i := 0; i < 3; i++ {
		teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
		policyWithIds := createRetentionPolicyWithTeamAndChannelIds(
			"Policy "+strconv.Itoa(i+1), teamIDs, channelIDs)
		policyWithCounts, err := ss.RetentionPolicy().Save(policyWithIds)
		require.Nil(t, err)
		policiesWithCounts = append(policiesWithCounts, policyWithCounts)
	}

	t.Run("get all", func(t *testing.T) {
		retrievedPolicies, err := ss.RetentionPolicy().GetAll(0, 60)
		require.Nil(t, err)
		require.Equal(t, len(policiesWithCounts), len(retrievedPolicies))
		for i := range policiesWithCounts {
			CheckRetentionPolicyWithTeamAndChannelCountsAreEqual(t, policiesWithCounts[i], retrievedPolicies[i])
		}
	})
	t.Run("get all with limit", func(t *testing.T) {
		for i := range policiesWithCounts {
			retrievedPolicies, err := ss.RetentionPolicy().GetAll(i, 1)
			require.Nil(t, err)
			require.Equal(t, 1, len(retrievedPolicies))
			CheckRetentionPolicyWithTeamAndChannelCountsAreEqual(t, policiesWithCounts[i], retrievedPolicies[0])
		}
	})
	t.Run("get all with same display name", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
			proposal := createRetentionPolicyWithTeamAndChannelIds(
				"Policy Name", teamIDs, channelIDs)
			_, err := ss.RetentionPolicy().Save(proposal)
			require.Nil(t, err)
		}
		policies, err := ss.RetentionPolicy().GetAll(0, 60)
		require.Nil(t, err)
		for i := 1; i < len(policies); i++ {
			require.True(t,
				policies[i-1].DisplayName < policies[i].DisplayName ||
					(policies[i-1].DisplayName == policies[i].DisplayName &&
						policies[i-1].ID < policies[i].ID),
				"policies with the same display name should be sorted by ID")
		}
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreGetCount(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("no policies", func(t *testing.T) {
		count, err := ss.RetentionPolicy().GetCount()
		require.NoError(t, err)
		require.Equal(t, int64(0), count)
	})
	t.Run("some policies", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy "+strconv.Itoa(i), nil, nil)
		}
		count, err := ss.RetentionPolicy().GetCount()
		require.NoError(t, err)
		require.Equal(t, int64(2), count)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreDelete(t *testing.T, ss store.Store, s SqlStore) {
	teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
	policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", teamIDs, channelIDs)
	t.Run("delete policy", func(t *testing.T) {
		err := ss.RetentionPolicy().Delete(policy.ID)
		require.Nil(t, err)
		policies, err := ss.RetentionPolicy().GetAll(0, 1)
		require.Nil(t, err)
		require.Empty(t, policies)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreGetChannels(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("no channels", func(t *testing.T) {
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", nil, nil)
		channels, err := ss.RetentionPolicy().GetChannels(policy.ID, 0, 1)
		require.Nil(t, err)
		require.Len(t, channels, 0)
	})
	t.Run("some channels", func(t *testing.T) {
		teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 2", teamIDs, channelIDs)
		channels, err := ss.RetentionPolicy().GetChannels(policy.ID, 0, len(channelIDs))
		require.Nil(t, err)
		require.Len(t, channels, len(channelIDs))
		sort.Strings(channelIDs)
		sort.Slice(channels, func(i, j int) bool {
			return channels[i].Id < channels[j].Id
		})
		for i := range channelIDs {
			require.Equal(t, channelIDs[i], channels[i].Id)
		}
	})
}

func testRetentionPolicyStoreAddChannels(t *testing.T, ss store.Store, s SqlStore) {
	teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
	policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", teamIDs, channelIDs)

	t.Run("add empty array", func(t *testing.T) {
		err := ss.RetentionPolicy().AddChannels(policy.ID, []string{})
		require.Nil(t, err)
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	t.Run("add new channels", func(t *testing.T) {
		channelIDs := createChannelsForRetentionPolicy(t, ss, teamIDs[0], 2)
		err := ss.RetentionPolicy().AddChannels(policy.ID, channelIDs)
		require.Nil(t, err)
		// verify that the channels were actually added
		copy := copyRetentionPolicyWithTeamAndChannelIds(policy)
		copy.ChannelIDs = append(copy.ChannelIDs, channelIDs...)
		checkRetentionPolicyLikeThisExists(t, ss, copy)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("add channel which does not exist", func(t *testing.T) {
		err := ss.RetentionPolicy().AddChannels(policy.ID, []string{"no_such_channel"})
		require.NotNil(t, err)
	})
	t.Run("add channel to policy which does not exist", func(t *testing.T) {
		channelIDs := createChannelsForRetentionPolicy(t, ss, teamIDs[0], 1)
		err := ss.RetentionPolicy().AddChannels("no_such_policy", channelIDs)
		require.NotNil(t, err)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreRemoveChannels(t *testing.T, ss store.Store, s SqlStore) {
	teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
	policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", teamIDs, channelIDs)

	t.Run("remove empty array", func(t *testing.T) {
		err := ss.RetentionPolicy().RemoveChannels(policy.ID, []string{})
		require.Nil(t, err)
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	t.Run("remove existing channel", func(t *testing.T) {
		channelID := channelIDs[0]
		err := ss.RetentionPolicy().RemoveChannels(policy.ID, []string{channelID})
		require.Nil(t, err)
		// verify that the channel was actually removed
		copy := copyRetentionPolicyWithTeamAndChannelIds(policy)
		copy.ChannelIDs = make([]string, 0)
		for _, oldChannelID := range policy.ChannelIDs {
			if oldChannelID != channelID {
				copy.ChannelIDs = append(copy.ChannelIDs, oldChannelID)
			}
		}
		checkRetentionPolicyLikeThisExists(t, ss, copy)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("remove channel which does not exist", func(t *testing.T) {
		err := ss.RetentionPolicy().RemoveChannels(policy.ID, []string{"no_such_channel"})
		require.Nil(t, err)
		// verify that the policy did not change
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreGetTeams(t *testing.T, ss store.Store, s SqlStore) {
	t.Run("no teams", func(t *testing.T) {
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", nil, nil)
		teams, err := ss.RetentionPolicy().GetTeams(policy.ID, 0, 1)
		require.Nil(t, err)
		require.Len(t, teams, 0)
	})
	t.Run("some teams", func(t *testing.T) {
		teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
		policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 2", teamIDs, channelIDs)
		teams, err := ss.RetentionPolicy().GetTeams(policy.ID, 0, len(teamIDs))
		require.Nil(t, err)
		require.Len(t, teams, len(teamIDs))
		sort.Strings(teamIDs)
		sort.Slice(teams, func(i, j int) bool {
			return teams[i].Id < teams[j].Id
		})
		for i := range teamIDs {
			require.Equal(t, teamIDs[i], teams[i].Id)
		}
	})
}

func testRetentionPolicyStoreAddTeams(t *testing.T, ss store.Store, s SqlStore) {
	teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
	policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", teamIDs, channelIDs)

	t.Run("add empty array", func(t *testing.T) {
		err := ss.RetentionPolicy().AddTeams(policy.ID, []string{})
		require.Nil(t, err)
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	t.Run("add new teams", func(t *testing.T) {
		teamIDs := createTeamsForRetentionPolicy(t, ss, 2)
		err := ss.RetentionPolicy().AddTeams(policy.ID, teamIDs)
		require.Nil(t, err)
		// verify that the teams were actually added
		copy := copyRetentionPolicyWithTeamAndChannelIds(policy)
		copy.TeamIDs = append(copy.TeamIDs, teamIDs...)
		checkRetentionPolicyLikeThisExists(t, ss, copy)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("add team which does not exist", func(t *testing.T) {
		err := ss.RetentionPolicy().AddTeams(policy.ID, []string{"no_such_team"})
		require.NotNil(t, err)
	})
	t.Run("add team to policy which does not exist", func(t *testing.T) {
		teamIDs := createTeamsForRetentionPolicy(t, ss, 1)
		err := ss.RetentionPolicy().AddTeams("no_such_policy", teamIDs)
		require.NotNil(t, err)
	})
	cleanupRetentionPolicyTest(s)
}

func testRetentionPolicyStoreRemoveTeams(t *testing.T, ss store.Store, s SqlStore) {
	teamIDs, channelIDs := createTeamsAndChannelsForRetentionPolicy(t, ss)
	policy := saveRetentionPolicyWithTeamAndChannelIds(t, ss, "Policy 1", teamIDs, channelIDs)

	t.Run("remove empty array", func(t *testing.T) {
		err := ss.RetentionPolicy().RemoveTeams(policy.ID, []string{})
		require.Nil(t, err)
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	t.Run("remove existing team", func(t *testing.T) {
		teamID := teamIDs[0]
		err := ss.RetentionPolicy().RemoveTeams(policy.ID, []string{teamID})
		require.Nil(t, err)
		// verify that the team was actually removed
		copy := copyRetentionPolicyWithTeamAndChannelIds(policy)
		copy.TeamIDs = make([]string, 0)
		for _, oldTeamID := range policy.TeamIDs {
			if oldTeamID != teamID {
				copy.TeamIDs = append(copy.TeamIDs, oldTeamID)
			}
		}
		checkRetentionPolicyLikeThisExists(t, ss, copy)
		restoreRetentionPolicy(t, ss, policy)
	})
	t.Run("remove team which does not exist", func(t *testing.T) {
		err := ss.RetentionPolicy().RemoveTeams(policy.ID, []string{"no_such_team"})
		require.Nil(t, err)
		// verify that the policy did not change
		checkRetentionPolicyLikeThisExists(t, ss, policy)
	})
	cleanupRetentionPolicyTest(s)
}