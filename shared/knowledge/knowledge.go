// Package knowledge provides known and commonly required information about the
// user to 3rd party apps, such as a user's last known location.
package knowledge

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/avabot/ava/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	"github.com/avabot/ava/shared/datatypes"
	"github.com/avabot/ava/shared/language"
)

var ErrNoLocation = errors.New("no previous location")

// GetLocation returns the last known location of a user. If the location isn't
// recent, ask the user to confirm.
func GetLocation(db *sqlx.DB, u *dt.User) (*dt.Location, string,
	error) {
	var loc *dt.Location
	if u.LocationID == 0 {
		return loc, language.QuestionLocation(""), nil
	}
	q := `
		SELECT name, createdat
		FROM locations
		WHERE userid=$1
		ORDER BY createdat DESC`
	err := db.Get(loc, q, u.ID)
	if err == sql.ErrNoRows {
		return loc, language.QuestionLocation(""), nil
	} else if err != nil {
		return loc, "", err
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	if loc.CreatedAt.Before(yesterday) {
		return loc, language.QuestionLocation(loc.Name), nil
	}
	return loc, "", nil
}

func GetAddress(db *sqlx.DB, u *dt.User, sent string) (*dt.Address, error) {
	var val string
	for _, w := range strings.Fields(sent) {
		if w == "home" || w == "office" {
			val = w
			break
		}
	}
	if len(val) == 0 {
		return nil, nil
	}
	q := `
		SELECT name, line1, line2, city, state, country, zip
		WHERE userid=$1 AND name=$2 AND cardid=0`
	var addr *dt.Address
	if err := db.Get(addr, q, u.ID, val); err != nil {
		return nil, err
	}
	return addr, nil
}
