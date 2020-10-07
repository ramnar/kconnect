/*
Copyright 2020 The kconnect Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package history

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/history/loader/mock_loader"
	"github.com/fidelity/kconnect/pkg/matchers"
)

func Test_FileStoreAdd(t *testing.T) {
	testCases := []struct {
		name            string
		input           *historyv1alpha.HistoryEntry
		existingHistory *historyv1alpha.HistoryEntryList
		maxItems        int
		expect          func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList)
		errorExpected   bool
	}{
		{
			name:            "Empty history, add entry",
			input:           historyv1alpha.NewHistoryEntry(),
			existingHistory: historyv1alpha.NewHistoryEntryList(),
			maxItems:        10,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				expectedList := historyv1alpha.NewHistoryEntryList()
				expectedList.Items = append(expectedList.Items, *input)

				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)

				mockLoader.
					EXPECT().
					Save(matchers.MatchHistoryList(expectedList)).
					DoAndReturn(func(historyList *historyv1alpha.HistoryEntryList) error {
						return nil
					}).Times(1)
			},
			errorExpected: false,
		},
		{
			name:            "Existing history below max items, add new entry",
			input:           createEntry("2"),
			existingHistory: createHistoryList(2),
			maxItems:        3,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				expectedList := historyv1alpha.NewHistoryEntryList()
				expectedList.Items = append(expectedList.Items, existing.Items...)
				expectedList.Items = append(expectedList.Items, *input)

				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)

				mockLoader.
					EXPECT().
					Save(matchers.MatchHistoryList(expectedList)).
					DoAndReturn(func(historyList *historyv1alpha.HistoryEntryList) error {
						return nil
					}).Times(1)
			},
			errorExpected: false,
		},
		{
			name:            "Existing history below max items, add entry for existing connection",
			input:           createEntry("0"),
			existingHistory: createHistoryList(1),
			maxItems:        3,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				expectedList := historyv1alpha.NewHistoryEntryList()
				expectedList.Items = existing.Items

				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)

				mockLoader.
					EXPECT().
					Save(matchers.MatchHistoryList(expectedList)).
					DoAndReturn(func(historyList *historyv1alpha.HistoryEntryList) error {
						return nil
					}).Times(1)
			},
			errorExpected: false,
		},
		{
			name:            "Existing history at max items, add new entry",
			input:           createEntry("2"),
			existingHistory: createHistoryList(2),
			maxItems:        2,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				expectedList := historyv1alpha.NewHistoryEntryList()
				entry2 := createEntry("1")
				expectedList.Items = append(expectedList.Items, *entry2)
				expectedList.Items = append(expectedList.Items, *input)

				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)

				mockLoader.
					EXPECT().
					Save(matchers.MatchHistoryList(expectedList)).
					DoAndReturn(func(historyList *historyv1alpha.HistoryEntryList) error {
						return nil
					}).Times(1)
			},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLoader := mock_loader.NewMockLoader(ctrl)

			tc.expect(mockLoader, tc.input, tc.existingHistory)

			store, err := NewStore(tc.maxItems, mockLoader)
			if err != nil {
				t.Fatalf("Failed to create history store: %v", err)
			}

			err = store.Add(tc.input)
			if tc.errorExpected && err == nil {
				t.Fatal("expected error on storing but not no error")
			}
			if !tc.errorExpected && err != nil {
				t.Fatalf("got an unexpected error: %v", err)
			}
		})
	}
}

func Test_FileStoreRemove(t *testing.T) {
	testCases := []struct {
		name            string
		input           *historyv1alpha.HistoryEntry
		existingHistory *historyv1alpha.HistoryEntryList
		maxItems        int
		expect          func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList)
		errorExpected   bool
	}{
		{
			name:            "Existing history, remove entry",
			input:           createEntry("1"),
			existingHistory: createHistoryList(2),
			maxItems:        10,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				expectedList := historyv1alpha.NewHistoryEntryList()
				entry := createEntry("0")
				expectedList.Items = append(expectedList.Items, *entry)

				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)

				mockLoader.
					EXPECT().
					Save(matchers.MatchHistoryList(expectedList)).
					DoAndReturn(func(historyList *historyv1alpha.HistoryEntryList) error {
						return nil
					}).Times(1)
			},
			errorExpected: false,
		},
		{
			name:            "No history, remove entry",
			input:           createEntry("1"),
			existingHistory: historyv1alpha.NewHistoryEntryList(),
			maxItems:        10,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)
			},
			errorExpected: true,
		},
		{
			name:            "Existing history, remove entry not in history",
			input:           createEntry("55"),
			existingHistory: createHistoryList(2),
			maxItems:        10,
			expect: func(mockLoader *mock_loader.MockLoader, input *historyv1alpha.HistoryEntry, existing *historyv1alpha.HistoryEntryList) {
				mockLoader.
					EXPECT().
					Load().
					DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
						return existing, nil
					}).Times(1)
			},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLoader := mock_loader.NewMockLoader(ctrl)

			tc.expect(mockLoader, tc.input, tc.existingHistory)

			store, err := NewStore(tc.maxItems, mockLoader)
			if err != nil {
				t.Fatalf("Failed to create history store: %v", err)
			}

			err = store.Remove([]*historyv1alpha.HistoryEntry{tc.input})
			if tc.errorExpected && err == nil {
				t.Fatal("expected error on storing but not no error")
			}
			if !tc.errorExpected && err != nil {
				t.Fatalf("got an unexpected error: %v", err)
			}
		})
	}
}

func Test_GetLastModified(t *testing.T) {
	testCases := []struct {
		name              string
		input             *historyv1alpha.HistoryEntryList
		expectedEntryName string
		errorExpected     bool
	}{
		{
			name:  "Single item",
			input: &historyv1alpha.HistoryEntryList{
				Items: []historyv1alpha.HistoryEntry {
					historyv1alpha.HistoryEntry{
						ObjectMeta: v1.ObjectMeta{
							Name: "test1",
						},
						Status: historyv1alpha.HistoryEntryStatus{
							LastUpdated: v1.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC),
						},
					},
				},
			},
			expectedEntryName: "test1",
			errorExpected: false,
		},
		{
			name:  "No items (error)",
			input: &historyv1alpha.HistoryEntryList{
				Items: []historyv1alpha.HistoryEntry {},
			},
			expectedEntryName: "",
			errorExpected: true,
		},
		{
			name:  "Multiple items",
			input: &historyv1alpha.HistoryEntryList{
				Items: []historyv1alpha.HistoryEntry {
					historyv1alpha.HistoryEntry{
						ObjectMeta: v1.ObjectMeta{
							Name: "test1",
						},
						Status: historyv1alpha.HistoryEntryStatus{
							LastUpdated: v1.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC),
						},
					},
					historyv1alpha.HistoryEntry{
						ObjectMeta: v1.ObjectMeta{
							Name: "test2",
						},
						Status: historyv1alpha.HistoryEntryStatus{
							LastUpdated: v1.Date(2020, 3, 1, 1, 1, 1, 1, time.UTC),
						},
					},
					historyv1alpha.HistoryEntry{
						ObjectMeta: v1.ObjectMeta{
							Name: "test3",
						},
						Status: historyv1alpha.HistoryEntryStatus{
							LastUpdated: v1.Date(2020, 2, 1, 1, 1, 1, 1, time.UTC),
						},
					},
				},
			},
			expectedEntryName: "test2",
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLoader, err := createStore(ctrl, tc.input)
			if err != nil {

			}
			actualLastModified, err := mockLoader.GetLastModified()
			if tc.errorExpected && err == nil {
				t.Fatal("expected error on getting last modified item but not no error")
			}
			if !tc.errorExpected && err != nil {
				t.Fatalf("got an unexpected error: %v", err)
			}
			if  !tc.errorExpected && tc.expectedEntryName != actualLastModified.GetName(){
				t.Fatalf("expected entry %v, but got %v", tc.expectedEntryName, actualLastModified.GetName())
			}
		})
	}
}

func createEntry(id string) *historyv1alpha.HistoryEntry {
	created, _ := time.Parse(time.RFC3339, "2020-09-0109T11:00:00+00:00")

	entry := historyv1alpha.NewHistoryEntry()
	entry.ObjectMeta.Name = id
	entry.ObjectMeta.CreationTimestamp = v1.Time{
		Time: created,
	}
	entry.Status.LastUpdated = v1.Time{
		Time: created,
	}
	entry.Spec = historyv1alpha.HistoryEntrySpec{
		ProviderID: id,
	}

	return entry
}

func createHistoryList(numEntries int) *historyv1alpha.HistoryEntryList {
	list := historyv1alpha.NewHistoryEntryList()

	for i := 0; i < numEntries; i++ {
		entry := createEntry(strconv.Itoa(i))
		list.Items = append(list.Items, *entry)
	}

	return list
}

func createStore(ctrl *gomock.Controller, entriesList *historyv1alpha.HistoryEntryList) (Store, error) {
	mockLoader := mock_loader.NewMockLoader(ctrl)
	store, err := NewStore(10, mockLoader)
	if err != nil {
		return nil, err
	}

	mockLoader.
	    EXPECT().
		Load().
		DoAndReturn(func() (*historyv1alpha.HistoryEntryList, error) {
			return entriesList, nil
		}).Times(1)
	
	return store, err
}