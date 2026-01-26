package git

import (
	"context"
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// PushToRemote pushes the current branch to the origin remote repository.
// It returns an error if there's no remote repository or if the push fails.
func PushToRemote(ctx context.Context) error {
	repo, err := gogit.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the current branch name
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	branchName := headRef.Name()
	if !branchName.IsBranch() {
		return fmt.Errorf("HEAD is not a branch")
	}

	// Get the origin remote
	remote, err := repo.Remote("origin")
	if err != nil {
		if err == gogit.ErrRemoteNotFound {
			return fmt.Errorf("no remote repository named 'origin' found")
		}
		return fmt.Errorf("failed to get remote: %w", err)
	}

	// Prepare the push reference: push current branch to remote branch with the same name
	refSpec := fmt.Sprintf("%s:%s", branchName.String(), branchName.String())

	// Execute the push
	err = remote.Push(&gogit.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(refSpec)},
		Progress:   nil,
	})

	// Handle common push errors
	if err != nil {
		if err == transport.ErrEmptyRemoteRepository {
			// First push to an empty repository is fine
			return nil
		}
		// Check if it's a "nothing to push" error (local and remote are the same)
		if err == transport.ErrAuthenticationRequired {
			return fmt.Errorf("authentication required for push: %w", err)
		}
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}
