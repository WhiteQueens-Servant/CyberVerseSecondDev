package scenic

import (
	"context"
	"time"
)

// DashboardData holds aggregated metrics for the admin dashboard.
type DashboardData struct {
	TodaySessions         int                    `json:"today_sessions"`
	WeekSessions          []DayCount             `json:"week_sessions"`
	HotQuestions          []QuestionCount        `json:"hot_questions"`
	SentimentDistribution SentimentDist          `json:"sentiment_distribution"`
	HourlyDistribution    []HourCount            `json:"hourly_distribution"`
}

// DayCount is a date with a count.
type DayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// QuestionCount is a question text with its frequency.
type QuestionCount struct {
	Question string `json:"question"`
	Count    int    `json:"count"`
}

// SentimentDist holds the breakdown of sentiment labels.
type SentimentDist struct {
	Positive int `json:"positive"`
	Neutral  int `json:"neutral"`
	Negative int `json:"negative"`
}

// HourCount is an hour (0-23) with a session count.
type HourCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// GetDashboard returns aggregated metrics for the admin dashboard.
func (d *DB) GetDashboard(ctx context.Context) (DashboardData, error) {
	var data DashboardData
	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	// 1. Today's session count
	err := d.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sessions WHERE started_at >= ?`,
		today+"T00:00:00Z").Scan(&data.TodaySessions)
	if err != nil {
		return data, err
	}

	// 2. Week sessions (last 7 days)
	weekAgo := now.AddDate(0, 0, -6).Format("2006-01-02")
	rows, err := d.conn.QueryContext(ctx,
		`SELECT SUBSTR(started_at, 1, 10) AS day, COUNT(*)
		 FROM sessions WHERE started_at >= ?
		 GROUP BY day ORDER BY day`,
		weekAgo+"T00:00:00Z")
	if err != nil {
		return data, err
	}
	defer rows.Close()
	// Build map for quick lookup
	dayMap := make(map[string]int)
	for rows.Next() {
		var dc DayCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return data, err
		}
		dayMap[dc.Date] = dc.Count
	}
	if err := rows.Err(); err != nil {
		return data, err
	}
	// Fill all 7 days (including zeros)
	data.WeekSessions = make([]DayCount, 7)
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -(6 - i)).Format("2006-01-02")
		data.WeekSessions[6-i] = DayCount{Date: day, Count: dayMap[day]}
	}

	// 3. Hot questions TOP 10 (group by normalized_content for better aggregation)
	rows2, err := d.conn.QueryContext(ctx,
		`SELECT normalized_content, COUNT(*) AS cnt
		 FROM conversation_logs
		 WHERE role = 'user' AND normalized_content != ''
		 GROUP BY normalized_content ORDER BY cnt DESC LIMIT 10`)
	if err != nil {
		return data, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var qc QuestionCount
		if err := rows2.Scan(&qc.Question, &qc.Count); err != nil {
			return data, err
		}
		data.HotQuestions = append(data.HotQuestions, qc)
	}
	if err := rows2.Err(); err != nil {
		return data, err
	}
	if data.HotQuestions == nil {
		data.HotQuestions = []QuestionCount{}
	}

	// 4. Sentiment distribution
	data.SentimentDistribution = SentimentDist{}
	rows3, err := d.conn.QueryContext(ctx,
		`SELECT sentiment, COUNT(*) FROM sessions GROUP BY sentiment`)
	if err != nil {
		return data, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var label string
		var count int
		if err := rows3.Scan(&label, &count); err != nil {
			return data, err
		}
		switch label {
		case "positive":
			data.SentimentDistribution.Positive = count
		case "negative":
			data.SentimentDistribution.Negative = count
		default:
			data.SentimentDistribution.Neutral += count
		}
	}
	if err := rows3.Err(); err != nil {
		return data, err
	}

	// 5. Hourly distribution (all time)
	rows4, err := d.conn.QueryContext(ctx,
		`SELECT CAST(SUBSTR(started_at, 12, 2) AS INTEGER) AS hour, COUNT(*)
		 FROM sessions GROUP BY hour ORDER BY hour`)
	if err != nil {
		return data, err
	}
	defer rows4.Close()
	hourMap := make(map[int]int)
	for rows4.Next() {
		var hc HourCount
		if err := rows4.Scan(&hc.Hour, &hc.Count); err != nil {
			return data, err
		}
		hourMap[hc.Hour] = hc.Count
	}
	if err := rows4.Err(); err != nil {
		return data, err
	}
	// Fill all 24 hours
	data.HourlyDistribution = make([]HourCount, 24)
	for h := 0; h < 24; h++ {
		data.HourlyDistribution[h] = HourCount{Hour: h, Count: hourMap[h]}
	}

	return data, nil
}
