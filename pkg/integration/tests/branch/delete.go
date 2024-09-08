package branch

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var Delete = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Try all combination of local and remote branch deletions",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.
			CloneIntoRemote("origin").
			EmptyCommit("blah").
			NewBranch("branch-one").
			EmptyCommit("on branch-one 01").
			PushBranchAndSetUpstream("origin", "branch-one").
			EmptyCommit("on branch-one 02").
			Checkout("master").
			Merge("branch-one"). // branch-one is contained in master, so no delete confirmation
			NewBranch("branch-two").
			EmptyCommit("on branch-two 01").
			PushBranchAndSetUpstream("origin", "branch-two"). // branch-two is contained in its own upstream, so no delete confirmation either
			NewBranchFrom("branch-three", "master").
			EmptyCommit("on branch-three 01").
			NewBranch("current-head"). // branch-three is contained in the current head, so no delete confirmation
			EmptyCommit("on current-head").
			NewBranchFrom("branch-four", "master").
			EmptyCommit("on branch-four 01").
			PushBranchAndSetUpstream("origin", "branch-four").
			EmptyCommit("on branch-four 02"). // branch-four is not contained in any of these, so we get a delete confirmation
			Checkout("current-head")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			Focus().
			Lines(
				Contains("current-head").IsSelected(),
				Contains("branch-four ↑1"),
				Contains("branch-three"),
				Contains("branch-two ✓"),
				Contains("master"),
				Contains("branch-one ↑1"),
			).

			// Deleting the current branch is not possible
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().
					Menu().
					Tooltip(Contains("You cannot delete the checked out branch!")).
					Title(Equals("Delete branch 'current-head'?")).
					Select(Contains("Delete local branch")).
					Confirm().
					Tap(func() {
						t.ExpectToast(Contains("You cannot delete the checked out branch!"))
					}).
					Cancel()
			}).

			// Delete branch-four. This is the only branch that is not fully merged, so we get
			// a confirmation popup.
			SelectNextItem().
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().
					Menu().
					Title(Equals("Delete branch 'branch-four'?")).
					Select(Contains("Delete local branch")).
					Confirm()
				t.ExpectPopup().
					Confirmation().
					Title(Equals("Force delete branch")).
					Content(Equals("'branch-four' is not fully merged. Are you sure you want to delete it?")).
					Confirm()
			}).
			Lines(
				Contains("current-head"),
				Contains("branch-three").IsSelected(),
				Contains("branch-two ✓"),
				Contains("master"),
				Contains("branch-one ↑1"),
			).

			// Delete branch-three. This branch is contained in the current head, so this just works
			// without any confirmation.
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().
					Menu().
					Title(Equals("Delete branch 'branch-three'?")).
					Select(Contains("Delete local branch")).
					Confirm()
			}).
			Lines(
				Contains("current-head"),
				Contains("branch-two ✓").IsSelected(),
				Contains("master"),
				Contains("branch-one ↑1"),
			).

			// Delete branch-two. This branch is contained in its own upstream, so this just works
			// without any confirmation.
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().
					Menu().
					Title(Equals("Delete branch 'branch-two'?")).
					Select(Contains("Delete local branch")).
					Confirm()
			}).
			Lines(
				Contains("current-head"),
				Contains("master").IsSelected(),
				Contains("branch-one ↑1"),
			).

			// Delete remote branch of branch-one. We only get the normal remote branch confirmation for this one.
			SelectNextItem().
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().
					Menu().
					Title(Equals("Delete branch 'branch-one'?")).
					Select(Contains("Delete remote branch")).
					Confirm()
			}).
			Tap(func() {
				t.ExpectPopup().
					Confirmation().
					Title(Equals("Delete branch 'branch-one'?")).
					Content(Equals("Are you sure you want to delete the remote branch 'branch-one' from 'origin'?")).
					Confirm()
			}).
			Tap(func() {
				t.Views().Remotes().
					Focus().
					Lines(Contains("origin")).
					PressEnter()

				t.Views().
					RemoteBranches().
					Lines(
						Equals("branch-four"),
						Equals("branch-two"),
					).
					Press(keys.Universal.Return)

				t.Views().
					Branches().
					Focus()
			}).
			Lines(
				Contains("current-head"),
				Contains("master"),
				Contains("branch-one (upstream gone)").IsSelected(),
			).

			// Delete local branch of branch-one. Even though its upstream is gone, we don't get a confirmation
			// because it is contained in master.
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().
					Menu().
					Title(Equals("Delete branch 'branch-one'?")).
					Select(Contains("Delete local branch")).
					Confirm()
			}).
			Lines(
				Contains("current-head"),
				Contains("master").IsSelected(),
			)
	},
})
