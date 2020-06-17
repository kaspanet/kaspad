package ff

import (
	"github.com/pkg/errors"
)

// rollback rolls the flat files on disk back to the provided file number
// and offset. This involves potentially deleting and truncating the files that
// were partially written.
//
// There are effectively two scenarios to consider here:
//   1) Transient write failures from which recovery is possible
//   2) More permanent failures such as hard disk death and/or removal
//
// In either case, the write cursor will be repositioned to the old flat file
// offset regardless of any other errors that occur while attempting to undo
// writes.
//
// For the first scenario, this will lead to any data which failed to be undone
// being overwritten and thus behaves as desired as the system continues to run.
//
// For the second scenario, the metadata which stores the current write cursor
// position within the flat files will not have been updated yet and thus if
// the system eventually recovers (perhaps the hard drive is reconnected), it
// will also lead to any data which failed to be undone being overwritten and
// thus behaves as desired.
func (s *flatFileStore) rollback(targetLocation *flatFileLocation) error {
	if s.isClosed {
		return errors.Errorf("cannot rollback a closed store %s",
			s.storeName)
	}

	// Grab the write cursor mutex since it is modified throughout this
	// function.
	s.writeCursor.Lock()
	defer s.writeCursor.Unlock()

	// Nothing to do if the rollback point is the same as the current write
	// cursor.
	targetFileNumber := targetLocation.fileNumber
	targetFileOffset := targetLocation.fileOffset
	if s.writeCursor.currentFileNumber == targetFileNumber && s.writeCursor.currentOffset == targetFileOffset {
		return nil
	}

	// If the rollback point is greater than the current write cursor then
	// something has gone very wrong, e.g. database corruption.
	if s.writeCursor.currentFileNumber < targetFileNumber ||
		(s.writeCursor.currentFileNumber == targetFileNumber && s.writeCursor.currentOffset < targetFileOffset) {
		return errors.Errorf("targetLocation is greater than the " +
			"current write cursor")
	}

	// Regardless of any failures that happen below, reposition the write
	// cursor to the target flat file and offset.
	defer func() {
		s.writeCursor.currentFileNumber = targetFileNumber
		s.writeCursor.currentOffset = targetFileOffset
	}()

	log.Warnf("ROLLBACK: Rolling back to file %d, offset %d",
		targetFileNumber, targetFileOffset)

	// Close the current write file if it needs to be deleted.
	if s.writeCursor.currentFileNumber > targetFileNumber {
		s.closeCurrentWriteCursorFile()
	}

	// Delete all files that are newer than the provided rollback file
	// while also moving the write cursor file backwards accordingly.
	s.lruMutex.Lock()
	defer s.lruMutex.Unlock()
	s.openFilesMutex.Lock()
	defer s.openFilesMutex.Unlock()
	for s.writeCursor.currentFileNumber > targetFileNumber {
		err := s.deleteFile(s.writeCursor.currentFileNumber)
		if err != nil {
			return errors.Wrapf(err, "ROLLBACK: Failed to delete file "+
				"number %d in store '%s'", s.writeCursor.currentFileNumber,
				s.storeName)
		}
		s.writeCursor.currentFileNumber--
	}

	// Open the file for the current write cursor if needed.
	s.writeCursor.currentFile.Lock()
	defer s.writeCursor.currentFile.Unlock()
	if s.writeCursor.currentFile.file == nil {
		openFile, err := s.openWriteFile(s.writeCursor.currentFileNumber)
		if err != nil {
			return err
		}
		s.writeCursor.currentFile.file = openFile
	}

	// Truncate the file to the provided target offset.
	err := s.writeCursor.currentFile.file.Truncate(int64(targetFileOffset))
	if err != nil {
		return errors.Wrapf(err, "ROLLBACK: Failed to truncate file %d "+
			"in store '%s'", s.writeCursor.currentFileNumber, s.storeName)
	}

	// Sync the file to disk.
	err = s.writeCursor.currentFile.file.Sync()
	if err != nil {
		return errors.Wrapf(err, "ROLLBACK: Failed to sync file %d in "+
			"store '%s'", s.writeCursor.currentFileNumber, s.storeName)
	}
	return nil
}
