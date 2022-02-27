package migration

import "sort"

func getMigrations(allMigrations []Migration, currentVersion int64) []Migration {
	sort.SliceStable(allMigrations, func(i, j int) bool {
		return allMigrations[i].Version < allMigrations[j].Version
	})

	relevantMigrations := []Migration{}

	for _, migration := range allMigrations {
		if migration.Version <= currentVersion {
			continue
		}

		relevantMigrations = append(relevantMigrations, migration)
	}

	return relevantMigrations
}
