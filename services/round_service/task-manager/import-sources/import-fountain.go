package importsources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FountainListSource struct {
	code string
	done bool
	// @Key: Page ID
	// @Value: Given Mark
	participants map[models.WikimediaUsernameType]struct{}
	jury         []models.WikimediaUsernameType
	json         *FountainJSON
}
type fountainMark struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Values []struct {
		Value       int    `json:"value"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Children    struct {
		} `json:"children"`
	} `json:"values"`
}

type GivenMarks struct {
	User    string         `json:"user"`
	Marks   map[string]int `json:"marks"`
	Comment string         `json:"comment"`
}
type fountainArticle struct {
	ID        uint64       `json:"id"`
	DateAdded time.Time    `json:"dateAdded"`
	Name      string       `json:"name"`
	User      string       `json:"user"`
	Marks     []GivenMarks `json:"marks"`
}
type fountainRule struct {
	Type   string         `json:"type"`
	Flags  int            `json:"flags"`
	Params map[string]any `json:"params"`
}
type FountainJSON struct {
	Code        string                  `json:"code"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Start       time.Time               `json:"start"`
	Finish      time.Time               `json:"finish"`
	Wiki        string                  `json:"wiki"`
	Flags       int                     `json:"flags"`
	Jury        []string                `json:"jury"`
	Marks       map[string]fountainMark `json:"marks"`
	MinMarks    int                     `json:"minMarks"`
	Rules       []fountainRule          `json:"rules"`
	Articles    []fountainArticle       `json:"articles"`
}

func (t *ImporterServer) ImportFromFountain(ctx context.Context, req *models.ImportFromFountainRequest) (*models.ImportResponse, error) {
	source := NewFountainListSource(req.Code)
	if source == nil {
		return nil, fmt.Errorf("FountainListSource is nil")
	}
	go t.importFrom(context.Background(), source, req.TaskId, req.RoundId)
	return &models.ImportResponse{}, nil
}
func NewFountainListSource(code string) *FountainListSource {
	return &FountainListSource{
		code: code,
	}
}
func (t *FountainListSource) parseFountain(reader io.Reader) []models.MediaResult {
	decoder := json.NewDecoder(reader)
	fountain := &FountainJSON{}
	// Here you would decode the JSON data into the fountain struct.
	// For example, using encoding/json package:
	err := decoder.Decode(fountain)
	if err != nil {
		// Handle error, e.g., log it or return it
		return nil
	}
	t.participants = map[models.WikimediaUsernameType]struct{}{}
	t.json = fountain
	t.jury = []models.WikimediaUsernameType{}
	for _, juryMember := range fountain.Jury {
		t.jury = append(t.jury, models.WikimediaUsernameType(juryMember))
	}
	data := []models.MediaResult{}
	// Process the fountain data as needed.
	for _, article := range t.json.Articles {
		pageId := article.ID
		data = append(data, models.MediaResult{
			PageID:              uint64(pageId),
			Name:                article.Name,
			SubmittedAt:         article.DateAdded,
			CreatedByUsername:   models.WikimediaUsernameType(article.User),
			SubmittedByUsername: models.WikimediaUsernameType(article.User),
			MediaType:           string(models.MediaTypeArticle),
		})
	}
	fmt.Printf("Fountain Code: %s, Name: %s, Description: %s\n", fountain.Code, fountain.Name, fountain.Description)
	fmt.Printf("Start: %s, Finish: %s, Wiki: %s\n", fountain.Start, fountain.Finish, fountain.Wiki)
	fmt.Printf("Flags: %d, MinMarks: %d\n", fountain.Flags, fountain.MinMarks)
	fmt.Printf("Jury: %v\n", fountain.Jury)
	return data
}
func (t *FountainListSource) ImportImageResults(ctx context.Context, currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	if t.done {
		return nil, nil
	}
	// This function is a placeholder for the actual import logic.
	// It should be implemented to handle the import of media results.
	// For now, it returns an empty slice and nil error.
	url := fmt.Sprintf("https://fountain.toolforge.org/api/editathons/%s", t.code)
	// Request the JSON data from the URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching fountain data: %v\n", err)
		return nil, &map[string]string{"error": "Failed to fetch fountain data"}
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching fountain data: received status code %d\n", resp.StatusCode)
		return nil, &map[string]string{"error": "Failed to fetch fountain data"}
	}
	data := t.parseFountain(resp.Body)
	t.done = true
	return data, nil
}

func (t *FountainListSource) PostProcess(ctx context.Context, tx *gorm.DB, currentRound *models.Round, task *models.Task, importMap map[uint64]types.SubmissionIDType, newlyCreatedUsers map[models.WikimediaUsernameType]models.IDType) error {
	fmt.Printf("Post-processing for round %s, task %s\n", currentRound.RoundID, task.TaskID)

	role_repo := repository.NewRoleRepository()
	juryRoleType := models.RoleTypeJury
	juryRoles, err := role_repo.FindRolesByUsername(tx.Preload("User"), t.jury, &models.RoleFilter{
		RoundID: &currentRound.RoundID,
		Type:    &juryRoleType,
	})
	if err != nil {
		return fmt.Errorf("failed to find jury roles: %w", err)
	}
	user_repo := repository.NewUserRepository()
	nameMap, err := user_repo.FetchExistingUsernames(tx, t.jury)
	if err != nil {
		return fmt.Errorf("failed to fetch existing usernames: %w", err)
	}
	roleMap := map[models.IDType]models.IDType{}
	for _, role := range juryRoles {
		roleMap[role.UserID] = role.RoleID
	}
	log.Printf("Found %d jury roles for fountain code %s\n", len(juryRoles), t.code)
	evaluations := []models.Evaluation{}
	markingPolicy := t.json.Marks
	for _, article := range t.json.Articles {
		pageId := article.ID
		uploaderId, ok := newlyCreatedUsers[models.WikimediaUsernameType(article.User)]
		if !ok {
			// If the user is not in newlyCreatedUsers, we might need to handle this
			log.Printf("User %s not found in newlyCreatedUsers, skipping evaluation for submission %d", article.User, pageId)
			continue
		}
		for _, evaluation := range article.Marks {
			actualMark := 0
			for markSegment, markIndex := range evaluation.Marks {
				actualMark = markingPolicy[markSegment].Values[markIndex].Value
			}
			submissionId := importMap[pageId]
			user := models.WikimediaUsernameType(evaluation.User)
			roleID := roleMap[nameMap[user]]
			score := models.ScoreType(float64(actualMark) * 100.0)
			now := time.Now()

			ev := models.Evaluation{
				SubmissionID:  submissionId,
				EvaluationID:  idgenerator.GenerateID("e"),
				JudgeID:       &roleID,
				RoundID:       currentRound.RoundID,
				Comment:       evaluation.Comment,
				Type:          models.EvaluationTypeScore,
				Score:         &score, // Assuming the score is a percentage
				EvaluatedAt:   &now,
				ParticipantID: uploaderId,
			}
			evaluations = append(evaluations, ev)
		}
	}
	if res := tx.Clauses(clause.Insert{Modifier: "IGNORE"}).CreateInBatches(evaluations, 1000); res.Error != nil {
		return fmt.Errorf("failed to create evaluations: %w", res.Error)
	}
	return nil
}
