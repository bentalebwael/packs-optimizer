package order_calculations

import (
	"context"
	"errors"
	"sort"

	"go.uber.org/zap"

	"github.com/pack-calculator/internal/pack_configurations"
	"github.com/pack-calculator/pkg/postgres"
)

type Service interface {
	OrderProcessing(ctx context.Context, orderQuantity int) (packs []PackResult, totalItems int, totalPacks int, err error)
	CalculateOptimalPacks(ctx context.Context, orderQuantity int, packSizes []int) (packCounts map[int]int, err error)
}

type service struct {
	logger          *zap.Logger
	calculationRepo Repository
	packsCfgRepo    pack_configurations.Repository
}

func NewService(logger *zap.Logger, calculationRepo Repository, packsCfgRepo pack_configurations.Repository) Service {
	return &service{
		logger:          logger,
		calculationRepo: calculationRepo,
		packsCfgRepo:    packsCfgRepo,
	}
}

func (s *service) OrderProcessing(ctx context.Context, orderQuantity int) ([]PackResult, int, int, error) {
	// Get available pack sizes
	packCfg, err := s.packsCfgRepo.GetActive(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	// Check if the calculation already exists in the database
	existingCalc, err := s.calculationRepo.GetByConfigurationIDAndOrderQuantity(ctx, orderQuantity, packCfg.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	if existingCalc != nil {
		s.logger.Info("Found existing calculation", zap.Int("orderQuantity", orderQuantity))
		return existingCalc.Result, existingCalc.TotalItems, existingCalc.TotalPacks, nil
	}

	packSizes := postgres.Int64ArrayToIntSlice(packCfg.PackSizes)
	if len(packSizes) == 0 {
		return nil, 0, 0, errors.New("no pack sizes available")
	}

	// Handle edge cases
	if orderQuantity == 0 || len(packSizes) == 0 {
		return []PackResult{}, 0, 0, nil
	}

	// Sort pack sizes to ensure we work from smallest to largest
	sort.Ints(packSizes)

	// Calculate optimal packs
	packCounts, err := s.CalculateOptimalPacks(ctx, orderQuantity, packSizes)
	if err != nil {
		return nil, 0, 0, err
	}

	var packs []PackResult
	totalItems := 0
	totalPacks := 0

	// Sort by pack size for consistent response
	sizes := make([]int, 0, len(packCounts))
	for size := range packCounts {
		sizes = append(sizes, size)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		quantity := packCounts[size]
		packs = append(packs, PackResult{
			Size:     size,
			Quantity: quantity,
		})
		totalItems += size * quantity
		totalPacks += quantity
	}

	// Save the order calculation to the database
	err = s.calculationRepo.Save(ctx, &OrderCalculation{
		OrderQuantity:   orderQuantity,
		TotalItems:      totalItems,
		TotalPacks:      totalPacks,
		Result:          packs,
		ConfigurationID: packCfg.ID,
	})
	if err != nil {
		return nil, 0, 0, err
	}

	return packs, totalItems, totalPacks, nil
}

// CalculateOptimalPacks finds the optimal combination of pack_configurations to fulfill an order
// Returns a map where keys are pack sizes and values are the number of pack_configurations needed
func (s *service) CalculateOptimalPacks(ctx context.Context, orderQuantity int, packSizes []int) (map[int]int, error) {
	// Fast-path for orders smaller than minimum pack size
	if orderQuantity <= packSizes[0] {
		return map[int]int{packSizes[0]: 1}, nil
	}

	// Step 1: Find the minimum total items needed to fulfill the order
	minTotal := findMinTotalItems(orderQuantity, packSizes)

	// Step 2: Find the minimal pack combination for this total
	packCounts := findMinPacks(minTotal, packSizes)

	return packCounts, nil
}

// findMinTotalItems finds the smallest possible total that can be created using
// available pack_configurations and is at least the order quantity
func findMinTotalItems(orderQuantity int, packSizes []int) int {
	// If no pack_configurations available, return -1 (error)
	if len(packSizes) == 0 {
		return -1
	}

	smallestPack := packSizes[0]
	maxPossibleTotal := orderQuantity + smallestPack - 1

	// dp[i] = true if we can make exactly i items using the available pack_configurations
	dp := make([]bool, maxPossibleTotal+1)
	dp[0] = true

	for _, packSize := range packSizes {
		for i := packSize; i <= maxPossibleTotal; i++ {
			if dp[i-packSize] {
				dp[i] = true
			}
		}
	}

	// Find the smallest valid total that's at least the order quantity
	for i := orderQuantity; i <= maxPossibleTotal; i++ {
		if dp[i] {
			return i
		}
	}

	return -1 // Should never happen if at least one pack size exists
}

// findMinPacks finds the minimum number of pack_configurations needed to make exactly the target total
func findMinPacks(targetTotal int, packSizes []int) map[int]int {
	if targetTotal <= 0 {
		return make(map[int]int)
	}

	// dp[i] = minimum number of pack_configurations needed to make i items
	dp := make([]int, targetTotal+1)
	for i := range dp {
		dp[i] = targetTotal + 1 // A value larger than any possible number of pack_configurations
	}
	dp[0] = 0

	// lastPack[i] = which pack size was last used to achieve total i
	lastPack := make([]int, targetTotal+1)

	for i := 1; i <= targetTotal; i++ {
		for _, packSize := range packSizes {
			if i >= packSize && dp[i-packSize] != targetTotal+1 {
				if dp[i-packSize]+1 < dp[i] {
					dp[i] = dp[i-packSize] + 1
					lastPack[i] = packSize
				}
			}
		}
	}

	// Reconstruct the solution
	packCounts := make(map[int]int)
	remainingTotal := targetTotal

	for remainingTotal > 0 {
		packSize := lastPack[remainingTotal]
		packCounts[packSize]++
		remainingTotal -= packSize
	}

	return packCounts
}
