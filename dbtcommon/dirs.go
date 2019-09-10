package dbtcommon

import (
	"fmt"
	"os"
	"path/filepath"
)

var pBits os.FileMode = 0755

// DirSpec holds the details for a hierarchy of directories
type DirSpec struct {
	name          string
	ignoreContent bool
	subDirs       []DirSpec
}

// Various names of directories and files
const (
	DbtDirName = "db.postgres"

	ReleaseScriptsBaseName  = "releaseScripts"
	ReleaseArchiveDirName   = "Archive"
	ReleaseSQLDirName       = "SQL.files"
	ReleaseManifestFileName = "Manifest"
	ReleaseReadMeFileName   = "ReadMe"
	ReleaseWarningFileName  = "Warning"

	MacrosDirName   = "macros"
	DatabaseDirName = "databases"

	SchemaDirName        = "schemas"
	SchemaSubDirTypes    = "types"
	SchemaSubDirTables   = "tables"
	SchemaSubDirFuncs    = "funcs"
	SchemaSubDirTriggers = "triggers"
)

var dirHierarchy = []DirSpec{
	{
		name: DbtDirName,
		subDirs: []DirSpec{
			{
				name: ReleaseScriptsBaseName,
				subDirs: []DirSpec{
					{
						name:          ReleaseArchiveDirName,
						ignoreContent: true,
					},
				},
			},
			{
				name:          MacrosDirName,
				ignoreContent: true,
			},
			{
				name:          DatabaseDirName,
				ignoreContent: true,
			},
		},
	},
}

var schemaDirs = []DirSpec{
	{
		name:          SchemaSubDirTypes,
		ignoreContent: true,
	},
	{
		name:          SchemaSubDirTables,
		ignoreContent: true,
	},
	{
		name:          SchemaSubDirFuncs,
		ignoreContent: true,
	},
	{
		name:          SchemaSubDirTriggers,
		ignoreContent: true,
	},
}

// DbtMacroDirName returns the name of the macros directory
func DbtMacroDirName() string {
	return filepath.Join(BaseDirName, DbtDirName, MacrosDirName)
}

// DbtDBBaseDirName returns the base name of the database directories
func DbtDBBaseDirName() string {
	return filepath.Join(BaseDirName, DbtDirName, DatabaseDirName)
}

// DbtDBDirName returns the name of the database directory
func DbtDBDirName(dbName string) string {
	return filepath.Join(DbtDBBaseDirName(), dbName)
}

// DbtSchemaBaseDirName returns the base name of the schema directories
func DbtSchemaBaseDirName(dbName string) string {
	return filepath.Join(DbtDBDirName(dbName), SchemaDirName)
}

// DbtSchemaDirName returns the full name of the directory for the given schema
func DbtSchemaDirName(dbName, schemaName string) string {
	return filepath.Join(DbtSchemaBaseDirName(dbName), schemaName)
}

// DbtReleaseBaseDirName returns the full name of the release scripts directory
func DbtReleaseBaseDirName() string {
	return filepath.Join(BaseDirName, DbtDirName, ReleaseScriptsBaseName)
}

// DbtReleaseDirName returns the full name of the release  directory
func DbtReleaseDirName(rel string) string {
	return filepath.Join(DbtReleaseBaseDirName(), rel)
}

// DbtReleaseSQLDirName returns the full name of the release SQL.files directory
func DbtReleaseSQLDirName(rel string) string {
	return filepath.Join(DbtReleaseBaseDirName(), rel, ReleaseSQLDirName)
}

// DbtReleaseManifestFile returns the full name of the release manifest file
func DbtReleaseManifestFile(rel string) string {
	return filepath.Join(DbtReleaseDirName(rel), ReleaseManifestFileName)
}

// DbtReleaseReadMeFile returns the full name of the release ReadMe file
func DbtReleaseReadMeFile(rel string) string {
	return filepath.Join(DbtReleaseDirName(rel), ReleaseReadMeFileName)
}

// DbtReleaseWarningFile returns the full name of the release Warning file
func DbtReleaseWarningFile(rel string) string {
	return filepath.Join(DbtReleaseDirName(rel), ReleaseWarningFileName)
}

// checkSubDirs recursively checks the dirs exist in base
func checkSubDirs(base string, dirs []DirSpec) bool {
	for _, d := range dirs {
		dirName := filepath.Join(base, d.name)
		info, err := os.Stat(dirName)
		if err != nil {
			return false
		}
		if !info.Mode().IsDir() {
			return false
		}
		if d.ignoreContent {
			continue
		}
		if !checkSubDirs(dirName, d.subDirs) {
			return false
		}
	}
	return true
}

// CheckDirs confirms that the necessary directories are present
func CheckDirs(dbName, schemaName string) bool {
	if !checkSubDirs(BaseDirName, dirHierarchy) {
		return false
	}

	return checkSubDirs(DbtSchemaDirName(dbName, schemaName), schemaDirs)
}

// makeDirIfMissing will create a directory if it is not present and will
// report any errors on the way
func makeDirIfMissing(dirName string) error {
	info, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		err = os.Mkdir(dirName, pBits)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !info.Mode().IsDir() {
		return fmt.Errorf("Couldn't create the directory %q"+
			" - it already exists and is not a directory",
			dirName)
	}
	return nil
}

// makeMissingSubDirs recursively makes the dirs missing from base. It stops
// at the first error
func makeMissingSubDirs(base string, dirs []DirSpec) error {
	for _, d := range dirs {
		dirName := filepath.Join(base, d.name)
		err := makeDirIfMissing(dirName)
		if err != nil {
			return err
		}

		if !d.ignoreContent {
			if err = makeMissingSubDirs(dirName, d.subDirs); err != nil {
				return err
			}
		}
	}
	return nil
}

// MakeMissingDirs this will make any directories that are needed
// and not present. There can be errors if the process doesn't have
// the right permissions, if there is a file-system object such as
// a file that is masking the directory, the file system is full
// etc. The attempt will stop at the first error
func MakeMissingDirs(dbName, schemaName string) error {
	err := makeMissingSubDirs(BaseDirName, dirHierarchy)
	if err != nil {
		return err
	}

	dirName := DbtDBDirName(dbName)
	err = makeDirIfMissing(dirName)
	if err != nil {
		return err
	}

	dirName = DbtSchemaBaseDirName(dbName)
	err = makeDirIfMissing(dirName)
	if err != nil {
		return err
	}

	dirName = DbtSchemaDirName(dbName, schemaName)
	err = makeDirIfMissing(dirName)
	if err != nil {
		return err
	}

	err = makeMissingSubDirs(dirName, schemaDirs)
	return err
}
