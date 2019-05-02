package main

import (
	"io"
	"errors"
	"html/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

func sessUserID(c echo.Context) int64 {
	sess, _ := session.Get("session", c)
	var userID int64
	if x, ok := sess.Values["user_id"]; ok {
		userID, _ = x.(int64)
	}
	return userID
}

func sessSetUserID(c echo.Context, id int64) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}
	sess.Values["user_id"] = id
	sess.Save(c.Request(), c.Response())
}

func sessDeleteUserID(c echo.Context) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}
	delete(sess.Values, "user_id")
	sess.Save(c.Request(), c.Response())
}

func sessAdministratorID(c echo.Context) int64 {
	sess, _ := session.Get("session", c)
	var administratorID int64
	if x, ok := sess.Values["administrator_id"]; ok {
		administratorID, _ = x.(int64)
	}
	return administratorID
}

func sessSetAdministratorID(c echo.Context, id int64) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}
	sess.Values["administrator_id"] = id
	sess.Save(c.Request(), c.Response())
}

func sessDeleteAdministratorID(c echo.Context) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}
	delete(sess.Values, "administrator_id")
	sess.Save(c.Request(), c.Response())
}

func loginRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := getLoginUser(c); err != nil {
			return resError(c, "login_required", 401)
		}
		return next(c)
	}
}

func adminLoginRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := getLoginAdministrator(c); err != nil {
			return resError(c, "admin_login_required", 401)
		}
		return next(c)
	}
}

func getLoginUser(c echo.Context) (*User, error) {
	userID := sessUserID(c)
	if userID == 0 {
		return nil, errors.New("not logged in")
	}
	var user User
	err := db.QueryRow("SELECT id, nickname FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Nickname)
	return &user, err
}

func getLoginAdministrator(c echo.Context) (*Administrator, error) {
	administratorID := sessAdministratorID(c)
	if administratorID == 0 {
		return nil, errors.New("not logged in")
	}
	var administrator Administrator
	err := db.QueryRow("SELECT id, nickname FROM administrators WHERE id = ?", administratorID).Scan(&administrator.ID, &administrator.Nickname)
	return &administrator, err
}

func getEvents(all bool) ([]*Event, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	rows, err := tx.Query("SELECT id FROM events ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event *Event
		if err := rows.Scan(event.ID); err != nil {
			return nil, err
		}
		event, err = getEventWithoutDetail(event.ID)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
	// var events []*Event
	// for rows.Next() {
	// 	var event Event
	// 	if err := rows.Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
	// 		return nil, err
	// 	}
	// 	if !all && !event.PublicFg {
	// 		continue
	// 	}
	// 	events = append(events, &event)
	// }

	// for i, v := range events {
	// 	event, err := getEvent(v.ID, -1)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	for k := range event.Sheets {
	// 		event.Sheets[k].Detail = nil
	// 	}
	// 	events[i] = event
	// }
	return events, nil
}


func getSheet(id int64) (*Sheet, error) {
	var s Sheet
	s.ID = id
	switch {
	case id <= 50:
		s.Num = id
		s.Price = 5000
		s.Rank = "S"
	case 50 < id && id <= 200:
		s.Num = id - 50
		s.Price = 3000
		s.Rank = "A"
	case 200 < id && id <= 500:
		s.Num = id - 200
		s.Price = 1000
		s.Rank = "B"
	case 500 < id && id <= 1000:
		s.Num = id - 500
		s.Price = 0
		s.Rank = "C"
	default:
		return nil, errors.New("invalid id error")
	}

	return &s, nil
}

func getRankAndNum(id int64) (string, int64) {
	var rank string
	var num int64
	switch {
	case id <= 50:
		rank, num = "S", id
	case 50 < id && id <= 200:
		rank, num = "A", id - 50
	case 200 < id && id <= 500:
		rank, num = "B", id - 200
	case 500 < id && id <= 1000:
		rank, num = "C", id - 500
	}
	return rank, num
}

func (e *Event) setSheetsWithoutDetail() {
	e.Total = 1000
	e.Remains = 1000
	e.Sheets = map[string]*Sheets{
		"S": &Sheets{Total: 50, Price: e.Price + 5000, Remains: 50},
		"A": &Sheets{Total: 150, Price: e.Price + 3000, Remains: 150},
		"B": &Sheets{Total: 300, Price: e.Price + 1000, Remains: 300},
		"C": &Sheets{Total: 500, Price: e.Price, Remains: 500},
	}
}

func (e *Event) setSheets() error {
	e.setSheetsWithoutDetail()
	var err error
	var sheet *Sheet
	for i:= 0; i < 1000; i++ {
		sheet, err = getSheet(int64(i+1))
		e.Sheets[sheet.Rank].Detail = append(e.Sheets[sheet.Rank].Detail, sheet)
	}
	return err
}

func getEventWithoutDetail(eventID int64) (*Event, error) {
	var event Event
	if err := db.QueryRow("SELECT * FROM events WHERE id = ?", eventID).Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
		return nil, err
	}
	
	event.setSheetsWithoutDetail()
	

	rows, err := db.Query("SELECT r.user_id, r.sheet_id, r.reserved_at FROM reservations r WHERE r.event_id = ? AND r.canceled_at IS NULL", event.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var reservation Reservation
		err := rows.Scan(&reservation.UserID, &reservation.SheetID, &reservation.ReservedAt)

		if err == nil {
			rank, _ := getRankAndNum(reservation.SheetID)
			event.Sheets[rank].Remains--
			
			event.Remains--
		} else {
			return nil, err
		}
	}

	return &event, nil
}

func getEvent(eventID, loginUserID int64) (*Event, error) {
	var event Event
	if err := db.QueryRow("SELECT * FROM events WHERE id = ?", eventID).Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
		return nil, err
	}
	
	err := event.setSheets()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT r.user_id, r.sheet_id, r.reserved_at FROM reservations r WHERE r.event_id = ? AND r.canceled_at IS NULL", event.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var reservation Reservation
		err := rows.Scan(&reservation.UserID, &reservation.SheetID, &reservation.ReservedAt)

		if err == nil {
			rank, num := getRankAndNum(reservation.SheetID)
			event.Sheets[rank].Detail[num-1].Mine = reservation.UserID == loginUserID
			event.Sheets[rank].Detail[num-1].Reserved = true
			event.Sheets[rank].Detail[num-1].ReservedAtUnix = reservation.ReservedAt.Unix()
			event.Sheets[rank].Remains--
			
			event.Remains--
		} else {
			return nil, err
		}
	}

	return &event, nil
}

func sanitizeEvent(e *Event) *Event {
	sanitized := *e
	sanitized.Price = 0
	sanitized.PublicFg = false
	sanitized.ClosedFg = false
	return &sanitized
}

func fillinUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if user, err := getLoginUser(c); err == nil {
			c.Set("user", user)
		}
		return next(c)
	}
}

func fillinAdministrator(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if administrator, err := getLoginAdministrator(c); err == nil {
			c.Set("administrator", administrator)
		}
		return next(c)
	}
}

func validateRank(rank string) bool {
	switch rank {
	case "S", "A", "B", "C":
		return true
	default:
		return false
	}
}

type Renderer struct {
	templates *template.Template
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}
