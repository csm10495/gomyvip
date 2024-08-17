package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sort"

	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sync/errgroup"
)

const BASE_URL = "https://loyalty-award-api.myvip.co/api/proxy/rewards/section/"

// RewardDataSimplified is a simplified version of the RewardData struct containing only the fields we are interested in.
type RewardDataSimplified struct {
	Name        string
	Price       int
	Description string
	Stock       int
	Partner     string
}

// RewardData is the struct that represents a part of the JSON response from the API.
type RewardData struct {
	Meta struct {
		Order          int    `json:"Order"`
		Type           string `json:"Type"`
		Featured       bool   `json:"Featured"`
		Title          string `json:"Title"`
		OccupiedSlots  int    `json:"OccupiedSlots"`
		AvailableSlots int    `json:"AvailableSlots"`
		Description    any    `json:"Description"`
		ImageURL       any    `json:"ImageURL"`
		HeroImageURL   any    `json:"HeroImageURL"`
		ID             any    `json:"_id"`
		IsCustom       bool   `json:"IsCustom"`
		IconURL        any    `json:"IconURL"`
	} `json:"Meta"`
	Awards []struct {
		AwardID          int    `json:"AwardID"`
		OfferID          int    `json:"OfferID"`
		TypeID           int    `json:"TypeID"`
		TypeSortOrder    int    `json:"TypeSortOrder"`
		PartnerID        int    `json:"PartnerId"`
		PropertyID       int    `json:"PropertyId"`
		PartnerName      string `json:"PartnerName"`
		Title            string `json:"Title"`
		ShortDescription string `json:"ShortDescription"`
		SubTitle         string `json:"SubTitle"`
		SubTitle2        any    `json:"SubTitle2"`
		SnipeText        string `json:"SnipeText"`
		SnipeCategory    string `json:"SnipeCategory"`
		ImageURL         string `json:"ImageURL"`
		Featured         bool   `json:"Featured"`
		Quantity         int    `json:"Quantity"`
		Price            int    `json:"Price"`
		LocationName     string `json:"LocationName"`
		Duration         int    `json:"Duration"`
		PlayerLimit      int    `json:"PlayerLimit"`
		UnlockLevel      int    `json:"UnlockLevel"`
		ExpireTime       string `json:"ExpireTime"`
		RequiredInfo     struct {
			Address bool `json:"Address"`
			Email   bool `json:"Email"`
		} `json:"RequiredInfo"`
		PriceOverride                any    `json:"PriceOverride"`
		OutletName                   string `json:"OutletName"`
		PropertyName                 string `json:"PropertyName"`
		DestinationID                any    `json:"DestinationId"`
		CollectionID                 any    `json:"CollectionId"`
		LoyaltyProgramName           any    `json:"LoyaltyProgramName"`
		AllowAutoRedeem              bool   `json:"AllowAutoRedeem"`
		RewardGiveAwayType           any    `json:"RewardGiveAwayType"`
		IsGiveAway                   bool   `json:"IsGiveAway"`
		IgnorePartnerRedemptionRules bool   `json:"IgnorePartnerRedemptionRules"`
		CanShowToUnqualified         bool   `json:"CanShowToUnqualified"`
		MinVipTier                   any    `json:"MinVipTier"`
		MaxmyVipTierLevelID          any    `json:"MaxmyVipTierLevelId"`
		RedemptionType               any    `json:"RedemptionType"`
		SortOrder                    int    `json:"SortOrder"`
		IsFavorite                   bool   `json:"IsFavorite"`
		StrikeOutPrice               any    `json:"StrikeOutPrice"`
		StrikeOutReason              string `json:"StrikeOutReason"`
		ForwardLink                  any    `json:"ForwardLink"`
		GalleryImageURL              string `json:"GalleryImageURL"`
		IsPremium                    bool   `json:"IsPremium"`
	} `json:"Awards"`
	Sections any `json:"Sections"`
}

// ToSimplified converts a RewardData struct to a RewardDataSimplified struct via coercing the fields a bit.
func (d RewardData) ToSimplified() mapset.Set[RewardDataSimplified] {
	simplified := mapset.NewSet[RewardDataSimplified]()

	for _, award := range d.Awards {
		partner := ""
		if award.LocationName != "" {
			partner = award.LocationName
		} else if award.PropertyName != "" {
			partner = award.PropertyName
		} else if award.PartnerName != "" {
			partner = award.PartnerName
		} else if award.OutletName != "" {
			partner = award.OutletName
		}

		quantity := -1
		if award.Quantity >= 0 {
			quantity = award.Quantity
		}

		description := ""
		if award.ShortDescription != "" {
			description = award.ShortDescription
		} else if award.SubTitle != "" {
			description = award.SubTitle
		}

		one := RewardDataSimplified{
			Name:        strings.TrimSpace(award.Title),
			Price:       award.Price,
			Description: strings.TrimSpace(description),
			Stock:       quantity,
			Partner:     partner,
		}
		simplified.Add(one)
	}

	return simplified
}

// RewardDataWrapper is the struct that represents the jull JSON response from the API.
type RewardDataWrapper struct {
	Meta struct {
		Order        int    `json:"Order"`
		Title        string `json:"Title"`
		SubTitle     string `json:"SubTitle"`
		Description  string `json:"Description"`
		HeroImageURL string `json:"HeroImageURL"`
		LogoImageURL string `json:"LogoImageURL"`
		CellImageURL string `json:"CellImageURL"`
	} `json:"Meta"`
	Lanes        []RewardData
	Message      any `json:"Message"`
	ErrorMessage any `json:"ErrorMessage"`
}

// DoGet makes a GET request to the API and returns a set of RewardDataSimplified structs.
func DoGet(slug string, page int) mapset.Set[RewardDataSimplified] {
	rewardData := mapset.NewSet[RewardDataSimplified]()

	url := BASE_URL + slug + "/" + fmt.Sprintf("%d", page)

	resp, err := http.Get(url)

	if err != nil {
		fmt.Printf("Error making GET request: %v\n", err)
		return rewardData
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// not logging anything here but technically this could be an error on our part.
		// Normally we assume instead that a non-OK means the API doesn't have a matching page.
		return rewardData
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return rewardData
	}

	wrapper := &RewardDataWrapper{}
	if err := json.Unmarshal(body, wrapper); err != nil {
		fmt.Printf("Error unmarshalling response body: %v\n", err)
		return rewardData
	}

	for lane := range wrapper.Lanes {
		rewardData = rewardData.Union(wrapper.Lanes[lane].ToSimplified())
	}
	return rewardData
}

func main() {
	allRewards := mapset.NewSet[RewardDataSimplified]()

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(4)

	// There is also a destination slug.. though it seems to just get dups.
	slugs := []string{"category"}
	for _, slug := range slugs {
		for idx := 0; idx <= 50; idx++ {
			i := idx
			s := slug
			g.Go(func() error {
				allRewards = allRewards.Union(DoGet(s, i))
				return nil
			})

		}
	}

	g.Wait()

	allRewardsSlice := allRewards.ToSlice()

	// Sort by price
	sort.Slice(allRewardsSlice, func(i, j int) bool {
		if allRewardsSlice[i].Price == allRewardsSlice[j].Price {
			if allRewardsSlice[i].Name == allRewardsSlice[j].Name {
				return allRewardsSlice[i].Partner < allRewardsSlice[j].Partner
			}
			return allRewardsSlice[i].Name < allRewardsSlice[j].Name
		}
		return allRewardsSlice[i].Price < allRewardsSlice[j].Price
	})

	s, _ := json.MarshalIndent(allRewardsSlice, "", "    ")
	fmt.Print(string(s))
}
