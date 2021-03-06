/*
   Written by Bradley Stanley-Clamp (bradley.stanley-clamp19@imperial.ac.uk) and Nicholas Pfaff (nicholas.pfaff19@imperial.ac.uk), 2021 - SpaceX++ EEE/EIE 2nd year group project, Imperial College London
*/

package server

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

func (s *SQLiteDB) getLogger() *zap.Logger {
	return s.logger
}

func (s *SQLiteDB) saveMapName(ctx context.Context, name string) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO maps (name)
			VALUES (:name)
		`,
			sql.Named("name", name),
		); err != nil {
			fmt.Println("not inserted:", name)
			return fmt.Errorf("server: SQLdb: failed to insert map name data into db: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: saveMapName transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) saveRover(ctx context.Context, mapID int, roverIndex int) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO rover (mapID, indx, rotation)
			VALUES (:mapID, :indx, :rotation)
		`,
			sql.Named("mapID", mapID),
			sql.Named("indx", roverIndex),
			sql.Named("rotation", Rover.Rotation),
		); err != nil {
			fmt.Println("rover not inserted")
			return fmt.Errorf("server: SQLdb: failed to insert rover data into db: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: saveRover transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertMap(ctx context.Context, tiles []int, mapID int) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {

		for i := 0; i < 144; i++ {

			if _, err := tx.ExecContext(ctx, `
			INSERT INTO tiles (indx, mapID, value)
			VALUES (:indx, :mapID, :value )
		`,
				sql.Named("indx", i),
				sql.Named("mapID", mapID),
				sql.Named("value", tiles[i]),
			); err != nil {
				fmt.Println("not inserted:", i, tiles[i], mapID)
				return fmt.Errorf("server: SQLdb: failed to insert map into db: %w", err)
			}
		}
		return nil

	}); err != nil {
		return fmt.Errorf("server: SQLdb: insertMap transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) retriveMap(ctx context.Context, mapID int) error {

	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT indx, value 
			FROM tiles 
			WHERE mapID = :mID
			`,
			sql.Named("mID", mapID),
		)
		if err != nil {
			return fmt.Errorf("server: SQLdb: failed to retrieve data from tiles rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tileID, value int
			if err := rows.Scan(
				&tileID,
				&value,
			); err != nil {
				return fmt.Errorf("server: SQLdb: failed to scan tiles row: %w", err)
			}
			dbMap.Tiles[tileID] = value
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("server: SQLdb: failed to scan last tile row: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: retriveMap transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) retriveRover(ctx context.Context, mapID int) error {
	var indx, rotation int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT indx, rotation
			FROM rover
			WHERE mapID = :mID
		`,
			sql.Named("mID", mapID),
		).Scan(&indx, &rotation); err != nil {
			return fmt.Errorf("server: SQLdb: failed to find id row: %w", err)
		}

		fmt.Println("rover index: ", indx, "rotation: ", rotation)
		dbMap.RoverRotation = rotation
		dbMap.RoverIndx = indx
		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: retriveRover transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) getMapID(ctx context.Context, name string) (int, error) {

	var id int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT mapID
			FROM maps
			WHERE name = :mapName
		`,
			sql.Named("mapName", name),
		).Scan(&id); err != nil {
			return fmt.Errorf("server: SQLdb: failed to find id row: %w", err)
		}
		return nil
	}); err != nil {
		return -1, fmt.Errorf("server: SQLdb: get map id transaction failed: %w", err)
	}

	return id, nil
}

func (s *SQLiteDB) getLatestMapID(ctx context.Context) (int, error) {

	var id int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT max(mapID)
			FROM maps
		`,
		).Scan(&id); err != nil {
			id = 0
			s.logger.Info("server: SQLdb: couldnt get latest map id, setting mapID to 0 (so next is 1)")

			return fmt.Errorf("server: SQLdb: failed to find id row: %w", err)
		}

		return nil
	}); err != nil {
		return -1, fmt.Errorf("server: SQLdb: getLatestMapID transaction failed: %w", err)
	}

	return id, nil
}

func (s *SQLiteDB) storeInstruction(ctx context.Context, instruction string, value int) error {
	s.logger.Info("storing instruction: inside function")

	mapID, err := s.getLatestMapID(ctx)
	if err != nil {
		fmt.Println("no mapID : ", mapID)
		fmt.Println("Error: couldnt get latest map ID")
	}

	s.logger.Info("storing instruction: map id", zap.Int("mapID", mapID))

	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO instructions (mapID, instruction, value)
			VALUES (:mapID, :instruction, :value )
		`,
			sql.Named("mapID", (mapID+1)),
			sql.Named("instruction", instruction),
			sql.Named("value", value),
		); err != nil {
			fmt.Println("not inserted:", mapID, instruction, value)
			return fmt.Errorf("server: SQLdb: failed to insert instruction into db: %w", err)
		}
		s.logger.Info("inserted instruction", zap.Int("mapID", mapID), zap.String("instruction", instruction), zap.Int("value", value))

		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: storeInstruction transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) retriveInstruction(ctx context.Context, mapID int) error {

	var instr string
	var val int

	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT instruction, value 
			FROM instructions
			WHERE mapID = :mID
			ORDER BY instructionID 
			`,
			sql.Named("mID", mapID),
		)
		if err != nil {
			fmt.Println("did not work extracting instructions")
			return fmt.Errorf("server: SQLdb: failed to retrieve data from instruction rows: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(
				&instr,
				&val,
			); err != nil {
				return fmt.Errorf("server: SQLdb: failed to scan instruction row: %w", err)
			}

			dbMap.Instructions = append(dbMap.Instructions, driveInstruction{
				Instruction: instr,
				Value:       val,
			})

		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("server: SQLdb: failed to scan last instruction row: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: retriveInstruction transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) resetInstructions(ctx context.Context, mapID int) error {

	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM instructions
			WHERE mapID = :mapID
		`,
			sql.Named("mapID", mapID),
		); err != nil {
			fmt.Println("instructions not deleted")
			return fmt.Errorf("server: SQLdb: failed to delete instructions from db: %w", err)
		}
		s.logger.Info("instructions deleted successfully")

		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: resetInstructions transaction failed: %w", err)
	}

	return nil

}

func (s *SQLiteDB) insertCredentials(ctx context.Context, credential credential) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO credentials (username, password)
			VALUES (:username, :password)
		`,
			sql.Named("username", credential.username),
			sql.Named("password", credential.password),
		); err != nil {
			return fmt.Errorf("server: SQLdb: failed to insert credential into db: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("server: SQLdb: insertCredentials transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) getCredentials(ctx context.Context) (map[string]string, error) {
	credentials := map[string]string{}
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT username, password
			FROM credentials
		`)
		if err != nil {
			return fmt.Errorf("server: SQLdb: failed to retrieve credential rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var username, password string
			if err := rows.Scan(
				&username,
				&password,
			); err != nil {
				return fmt.Errorf("server: SQLdb: failed to scan credential row: %w", err)
			}

			credentials[username] = password
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("server: SQLdb: failed to scan last credential row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("server: SQLdb: getCredentials transaction failed: %w", err)
	}

	return credentials, nil
}

func (s *SQLiteDB) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("server: SQLdb: failed to close sqlite db: %w", err)
	}
	return nil
}
