package repositories

import (
	"errors"
	"regexp"

	"github.com/FeaturePlus/backend/models"
	"gorm.io/gorm"
)

type ReleaseRepository interface {
	Create(release *models.Release, prIDs []int) error
	GetByID(id uint) (*models.Release, error)
	GetByTag(projectID uint, tag string) (*models.Release, error)
	PRsExist(prIDs []int) (bool, error)
	GetAll() ([]models.Release, error)
	UpdateStatus(id uint, status models.ReleaseStatus) error
	CheckPRsSameProject(prIDs []int) (bool, error)
	CheckPRsNotInOtherReleases(releaseID uint, prIDs []int) (bool, []int, error)
	DB() *gorm.DB
}

type releaseRepository struct {
	db *gorm.DB
}

func NewReleaseRepository(db *gorm.DB) ReleaseRepository {
	return &releaseRepository{db: db}
}

// DB returns the underlying *gorm.DB (internal helper for handlers needing ad-hoc queries)
func (r *releaseRepository) DB() *gorm.DB {
	return r.db
}

func (r *releaseRepository) Create(release *models.Release, prIDs []int) error {
	// Start a transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the release
	if err := tx.Create(release).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Associate PRs with the release
	for _, prID := range prIDs {
		if err := tx.Exec("INSERT INTO release_prs (release_id, pull_request_id) VALUES (?, ?)", release.ID, prID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *releaseRepository) GetByID(id uint) (*models.Release, error) {
	// First, fetch the release without PRs
	var release models.Release
	if err := r.db.First(&release, id).Error; err != nil {
		return nil, err
	}
	
	// Then, fetch the associated PRs using a join query to ensure we get the correct PRs
	var prs []models.PullRequest
	if err := r.db.Joins("JOIN release_prs ON release_prs.pull_request_id = pull_requests.id").Where("release_prs.release_id = ?", id).Find(&prs).Error; err != nil {
		return nil, err
	}
	
	// Assign the PRs to the release
	release.PRs = prs
	
	// Debug logging to verify which PRs were loaded
	var prIDs []uint
	for _, pr := range release.PRs {
		prIDs = append(prIDs, pr.ID)
	}
	r.db.Logger.Info(r.db.Statement.Context, "Loaded PRs for release %d: %v", id, prIDs)
	
	return &release, nil
}

func (r *releaseRepository) GetByTag(projectID uint, tag string) (*models.Release, error) {
	var release models.Release
	if err := r.db.Where("project_id = ? AND tag = ?", projectID, tag).First(&release).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &release, nil
}

func (r *releaseRepository) PRsExist(prIDs []int) (bool, error) {
	var count int64
	if err := r.db.Model(&models.PullRequest{}).Where("id IN ?", prIDs).Count(&count).Error; err != nil {
		return false, err
	}
	return int(count) == len(prIDs), nil
}

// ValidateTag checks if the tag follows semantic versioning (vX.Y.Z)
func ValidateTag(tag string) bool {
	re := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	return re.MatchString(tag)
}

func (r *releaseRepository) GetAll() ([]models.Release, error) {
	// First, fetch all releases without PRs
	var releases []models.Release
	if err := r.db.Find(&releases).Error; err != nil {
		return nil, err
	}
	
	// Then, for each release, fetch its associated PRs
	for i := range releases {
		var prs []models.PullRequest
		if err := r.db.Joins("JOIN release_prs ON release_prs.pull_request_id = pull_requests.id").Where("release_prs.release_id = ?", releases[i].ID).Find(&prs).Error; err != nil {
			return nil, err
		}
		
		// Assign the PRs to the release
		releases[i].PRs = prs
		
		// Debug logging to verify which PRs were loaded
		var prIDs []uint
		for _, pr := range prs {
			prIDs = append(prIDs, pr.ID)
		}
		r.db.Logger.Info(r.db.Statement.Context, "Loaded PRs for release %d: %v", releases[i].ID, prIDs)
	}
	
	return releases, nil
}

func (r *releaseRepository) UpdateStatus(id uint, status models.ReleaseStatus) error {
	return r.db.Model(&models.Release{}).Where("id = ?", id).Update("status", status).Error
}

// CheckPRsSameProject checks if all PRs belong to the same project
func (r *releaseRepository) CheckPRsSameProject(prIDs []int) (bool, error) {
	var prs []models.PullRequest
	if err := r.db.Where("id IN ?", prIDs).Find(&prs).Error; err != nil {
		return false, err
	}

	if len(prs) == 0 {
		return false, errors.New("no PRs found")
	}

	// Get the feature IDs for all PRs
	var featureIDs []uint
	for _, pr := range prs {
		featureIDs = append(featureIDs, pr.FeatureID)
	}

	// Get the projects for these features
	var features []models.Feature
	if err := r.db.Where("id IN ?", featureIDs).Find(&features).Error; err != nil {
		return false, err
	}

	if len(features) == 0 {
		return false, errors.New("no features found for the PRs")
	}

	// Check if all features belong to the same project
	projectID := features[0].ProjectID
	for _, feature := range features {
		if feature.ProjectID != projectID {
			return false, nil
		}
	}

	return true, nil
}

// CheckPRsNotInOtherReleases checks if any PRs are already part of another release
// Returns: (allPRsAvailable, listOfConflictingPRIDs, error)
func (r *releaseRepository) CheckPRsNotInOtherReleases(releaseID uint, prIDs []int) (bool, []int, error) {
	// Find PRs that are already in other releases (excluding the current release being finalized)
	var conflictingPRIDs []int

	// Query to find PRs that are in other releases
	rows, err := r.db.Raw(`
		SELECT rp.pull_request_id 
		FROM release_prs rp
		JOIN releases r ON rp.release_id = r.id
		WHERE rp.pull_request_id IN ? 
		AND rp.release_id != ? 
		AND r.status IN ('draft', 'published')
	`, prIDs, releaseID).Rows()

	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var prID int
		if err := rows.Scan(&prID); err != nil {
			return false, nil, err
		}
		conflictingPRIDs = append(conflictingPRIDs, prID)
	}

	return len(conflictingPRIDs) == 0, conflictingPRIDs, nil
}
