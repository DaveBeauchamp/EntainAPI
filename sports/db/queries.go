package db

const (
	eventList = "list"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventList: `
			SELECT 
				id, 
				name, 
				teamOne, 
				teamTwo, 
				visible, 
				advertised_start_time 
			FROM races
		`,
	}
}
