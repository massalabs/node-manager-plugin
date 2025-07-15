# Test Data Generator

This directory contains scripts to generate total value history test data for the node manager plugin.

## generate_test_data.go

A Go script that creates a test database with realistic value history data for testing the history retrieval functionality.

### Features

- Creates a SQLite database at `int/db/test_db/test.db`
- Populates the `value_history_buildnet` table with 5000 entries
- Spreads data over 1 year and 1 day (366 days)
- Uses realistic intervals and grouping patterns
- Generates entries with 3-minute intervals within groups
- Evenly distributes groups across the entire timespan (no random gaps)

### Data Characteristics

- **Total entries**: 5000
- **Time span**: 1 year and 1 day (366 days)
- **Interval**: 3 minutes between consecutive entries within groups
- **Group size**: Mean of 20 entries per group (with normal distribution)
- **Value range**: Starting at ~10,000 MAS with realistic variations
- **Distribution**: Groups are evenly spread across the timespan (no random gaps)

### Usage

```bash
task generate-test-data
```

