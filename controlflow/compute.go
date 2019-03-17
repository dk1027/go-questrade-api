package controlflow

import "log"

type Set map[string]struct{}
type Table map[string]float64

func Aggregate(mappings *map[string]string, portfolio *Portfolio) *Table {
	results := Table{}
	for _, p := range *portfolio {
		parent, ok := (*mappings)[p.Symbol]
		if !ok {
			log.Printf("Unknown mapping: %v. Ignored.\n", p.Symbol)
			continue
		}
		_, ok = results[parent]
		if !ok {
			results[parent] = 0
		}
		results[parent] += p.Amount
	}
	return &results
}

func toSet(ll *[]string) *Set {
	set := &Set{}
	for _, v := range *ll {
		(*set)[v] = struct{}{}
	}
	return set
}

// Filter filters out rows in portfolio containing ignoredSymbols / ignoredAccounts in-place
func Filter(ignoredSymbols *[]string, ignoredAccounts *[]string, portfolio *Portfolio) {
	accounts := toSet(ignoredAccounts)
	symbols := toSet(ignoredSymbols)
	i := 0
	for _, p := range *portfolio {
		if _, ignored := (*accounts)[p.Account]; ignored {
			continue
		}
		if _, ignored := (*symbols)[p.Symbol]; ignored {
			continue
		}
		(*portfolio)[i] = p
		i++
	}
	*portfolio = (*portfolio)[:i]
}

func CalculatePercentBalance(table *Table, targetAllocation *map[string]float64) (*Table, *Table) {
	var total float64
	percent := Table{}
	// Compute grant total
	for _, v := range *table {
		total += v
	}
	// Compute percentage of each group
	for k, v := range *table {
		percent[k] = v / total * 100
	}
	// Compute target amount
	target_amount := Table{}
	for k, v := range *targetAllocation {
		target_amount[k] = total * v
	}
	// Difference target - actual
	difference := Table{}
	for k, actual := range *table {
		t, _ := target_amount[k]
		difference[k] = t - actual
	}
	return &difference, &percent
}
