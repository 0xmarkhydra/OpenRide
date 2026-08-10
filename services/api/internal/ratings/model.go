package ratings

import "time"

type Rating struct {
	ID        string    `json:"id"`
	TripID    string    `json:"trip_id"`
	RiderID   string    `json:"rider_id"`
	DriverID  string    `json:"driver_id"`
	Stars     int16     `json:"stars"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
