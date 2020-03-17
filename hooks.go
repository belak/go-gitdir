package gitdir

import (
	"errors"
	"fmt"
	"io"

	"github.com/belak/go-gitdir/internal/git"
	"github.com/belak/go-gitdir/models"
)

// RunHook will run the given hook
func (c *Config) RunHook(
	hook string,
	repoPath string,
	pk *models.PublicKey,
	args []string,
	stdin io.Reader,
) error {
	user, err := c.LookupUserByKey(*pk, c.Options.GitUser)
	if err != nil {
		return err
	}

	repo, err := c.LookupRepoAccess(user, repoPath)
	if err != nil {
		return err
	}

	switch hook {
	case "pre-receive", "post-receive":
		// Pre and post are here just in case we need them in the future, but
		// they always succeed right now.
		return nil
	case "update":
		if len(args) < 3 {
			return errors.New("not enough args")
		}

		var (
			ref     = args[0]
			oldHash = args[1]
			newHash = args[2]
		)

		return runUpdateHook(repo, user, pk, git.NewHash(oldHash), git.NewHash(newHash), ref)
	default:
		return fmt.Errorf("hook %s is not implemented", hook)
	}
}

func runUpdateHook(
	lookup *RepoLookup,
	user *User,
	pk *models.PublicKey,
	oldHash git.Hash,
	newHash git.Hash,
	ref string,
) error {
	var (
		c   *Config
		err error
	)

	switch lookup.Type {
	case RepoTypeAdmin:
		c, err = LoadConfig("", newHash, nil, nil)
	case RepoTypeOrgConfig:
		c, err = LoadConfig("", git.ZeroHash, map[string]git.Hash{
			lookup.PathParts[0]: newHash,
		}, nil)
	case RepoTypeUserConfig:
		c, err = LoadConfig("", git.ZeroHash, nil, map[string]git.Hash{
			lookup.PathParts[0]: newHash,
		})
	default:
		// Non-admin repos don't need this hook.
		return nil
	}

	if err != nil {
		return err
	}

	return c.Validate(user, pk)
}