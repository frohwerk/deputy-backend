package test

import (
	"sort"
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/tree"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	tree.Log = logger.Basic(logger.LEVEL_DEBUG)

	db := DB()
	// db := database.Open()
	defer db.Close()

	exec := func(command string, args ...interface{}) {
		_, err := db.Exec(command, args...)
		if err != nil {
			t.Fatalf("error executing statement:\n%s\n%s\n", command, err)
		}
	}

	init := func() {
		exec(`DELETE FROM components WHERE char_length(id) = 6`)
		exec(`INSERT INTO components (id, name) values('top[0]','top tier component #1')`)
		exec(`INSERT INTO components (id, name) values('mid[0]','middle tier component #1')`)
		exec(`INSERT INTO components (id, name) values('mid[1]','middle tier component #2')`)
		exec(`INSERT INTO components (id, name) values('bot[0]','bottom tier component #1')`)
		exec(`INSERT INTO components (id, name) values('bot[1]','bottom tier component #2')`)
		exec(`INSERT INTO components (id, name) values('bot[2]','bottom tier component #3')`)
		exec(`INSERT INTO dependencies (id, depends_on) VALUES('top[0]', 'mid[0]')`)
		exec(`INSERT INTO dependencies (id, depends_on) VALUES('mid[0]', 'bot[1]')`)
		exec(`INSERT INTO dependencies (id, depends_on) VALUES('mid[0]', 'bot[2]')`)
	}

	lookup := func(id string) ([]string, error) {
		rows, err := db.Query(`SELECT depends_on FROM dependencies WHERE id = $1 ORDER BY depends_on`, id)
		if err != nil {
			return nil, err
		}
		result := []string{}
		for rows.Next() {
			var s string
			err := rows.Scan(&s)
			if err != nil {
				return nil, err
			}
			result = append(result, s)
		}
		return result, nil
	}

	init()

	t.Run("basic use case", func(t *testing.T) {
		if root, err := tree.Builder(lookup).CreateTree("top[0]"); assert.NoError(t, err) {
			// inspect initial result
			assert.Equal(t, "top[0]", root.Value)
			assert.Equal(t, 2, root.Depth)
			assert.Len(t, root.Dependencies, 1)
			mid := root.Dependencies
			assert.Equal(t, "mid[0]", mid[0].Value)
			assert.Equal(t, 1, mid[0].Depth)
			assert.Len(t, mid[0].Dependencies, 2)
			bot := mid[0].Dependencies
			assert.Equal(t, "bot[1]", bot[0].Value)
			assert.Len(t, bot[0].Dependencies, 0)
			assert.Equal(t, "bot[2]", bot[1].Value)
			assert.Len(t, bot[1].Dependencies, 0)
		}
	})

	t.Run("trimming", func(t *testing.T) {
		if root, err := tree.Builder(lookup).CreateTree("top[0]"); assert.NoError(t, err) {
			assert.False(t, root.Leaf())

			trimmed := root.Trim()
			assert.Len(t, trimmed, 2)
			sort.Slice(trimmed, func(i, j int) bool { return trimmed[i].Value < trimmed[j].Value })
			assert.Equal(t, "bot[1]", trimmed[0].Value)
			assert.Equal(t, "bot[2]", trimmed[1].Value)

			assert.Len(t, root.Dependencies, 1)
			mid := root.Dependencies
			assert.True(t, mid[0].Leaf(), "mid[0] should be a leaf now")
			assert.Equal(t, "mid[0]", mid[0].Value)

			trimmed = root.Trim()
			assert.Len(t, trimmed, 1)
			assert.Equal(t, "mid[0]", trimmed[0].Value)

			assert.True(t, root.Leaf(), "root should be a leaf now")
		}
	})

	t.Run("component without dependencies", func(t *testing.T) {
		if root, err := tree.Builder(lookup).CreateTree("mid[1]"); assert.NoError(t, err) {
			assert.Equal(t, "mid[1]", root.Value)
			assert.Len(t, root.Dependencies, 0)
		}
	})

	t.Run("caching", func(t *testing.T) {
		lookups := 0
		tb := tree.Builder(func(id string) ([]string, error) {
			lookups++
			return lookup(id)
		})
		tb.CreateTree("bot[1]") // Has no dependencies
		tb.CreateTree("bot[1]") // Has no dependencies
		assert.Equal(t, 1, lookups)
	})
}
