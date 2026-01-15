package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// NAV représente une valorisation (Net Asset Value) à une date donnée
type NAV struct {
	Date  string  // Format "2006-01-02"
	Value float64 // Valeur de la NAV
}

// Investment représente un investissement dans le portefeuille
type Investment struct {
	Name           string  // Nom de l'investissement
	AmountInvested float64 // Montant initial investi
	ReferenceRate  float64 // Taux de référence annuel (%)
	NAVHistory     []NAV   // Historique des NAV
	InvestmentDate string  // Date d'investissement initial
}

// Portfolio représente un portefeuille d'investissements
type Portfolio struct {
	Investments map[string]*Investment
}

// NewPortfolio crée un nouveau portefeuille vide
func NewPortfolio() *Portfolio {
	return &Portfolio{
		Investments: make(map[string]*Investment),
	}
}

// AddInvestment ajoute un nouvel investissement au portefeuille
func (p *Portfolio) AddInvestment(name string, amount float64, referenceRate float64, investmentDate string) error {
	if amount <= 0 {
		return fmt.Errorf("le montant doit être positif")
	}

	inv := &Investment{
		Name:           name,
		AmountInvested: amount,
		ReferenceRate:  referenceRate,
		NAVHistory:     make([]NAV, 0),
		InvestmentDate: investmentDate,
	}

	p.Investments[name] = inv
	return nil
}

// AddNAV ajoute une valorisation à un investissement
func (p *Portfolio) AddNAV(investmentName string, date string, value float64) error {
	inv, exists := p.Investments[investmentName]
	if !exists {
		return fmt.Errorf("l'investissement '%s' n'existe pas", investmentName)
	}

	if value <= 0 {
		return fmt.Errorf("la NAV doit être positive")
	}

	inv.NAVHistory = append(inv.NAVHistory, NAV{Date: date, Value: value})

	// Trier par date
	sort.Slice(inv.NAVHistory, func(i, j int) bool {
		return inv.NAVHistory[i].Date < inv.NAVHistory[j].Date
	})

	return nil
}

// GetLatestNAV retourne la dernière NAV connue pour un investissement
func (inv *Investment) GetLatestNAV() (NAV, error) {
	if len(inv.NAVHistory) == 0 {
		return NAV{}, fmt.Errorf("aucune NAV disponible")
	}
	return inv.NAVHistory[len(inv.NAVHistory)-1], nil
}

// CalculatePerformanceRate calcule le taux annuel de performance basé sur les données réelles
func (inv *Investment) CalculatePerformanceRate() (float64, error) {
	if len(inv.NAVHistory) < 2 {
		return 0, fmt.Errorf("au moins 2 NAV sont nécessaires")
	}

	firstNAV := inv.NAVHistory[0]
	lastNAV := inv.NAVHistory[len(inv.NAVHistory)-1]

	// Parser les dates
	t1, _ := time.Parse("2006-01-02", firstNAV.Date)
	t2, _ := time.Parse("2006-01-02", lastNAV.Date)

	years := t2.Sub(t1).Hours() / 24 / 365.25
	if years <= 0 {
		return 0, fmt.Errorf("l'intervalle de temps doit être positif")
	}

	// Formule: r = (VF/VI)^(1/n) - 1
	rate := math.Pow(lastNAV.Value/firstNAV.Value, 1/years) - 1
	return rate * 100, nil
}

// ProjectNAV projette la valeur future à une date donnée
func (inv *Investment) ProjectNAV(projectionDate string) (float64, error) {
	// Récupérer la dernière NAV connue
	latestNAV, err := inv.GetLatestNAV()
	if err != nil {
		return 0, err
	}

	// Calculer le taux de performance
	performanceRate := inv.ReferenceRate
	if len(inv.NAVHistory) >= 2 {
		calculatedRate, err := inv.CalculatePerformanceRate()
		if err == nil {
			// Prendre le taux le plus défavorable (le plus bas)
			if calculatedRate < performanceRate {
				performanceRate = calculatedRate
			}
		}
	}

	// Parser les dates
	t1, _ := time.Parse("2006-01-02", latestNAV.Date)
	t2, _ := time.Parse("2006-01-02", projectionDate)

	years := t2.Sub(t1).Hours() / 24 / 365.25
	if years < 0 {
		return 0, fmt.Errorf("la date de projection doit être après la dernière NAV")
	}

	// Formule: VF = VI * (1 + r)^n
	projectedValue := latestNAV.Value * math.Pow(1+(performanceRate/100), years)

	return projectedValue, nil
}

// GetPortfolioValue calcule la valeur totale du portefeuille à une date donnée
func (p *Portfolio) GetPortfolioValue(date string) (map[string]float64, float64, error) {
	values := make(map[string]float64)
	totalValue := 0.0

	for name, inv := range p.Investments {
		value, err := inv.ProjectNAV(date)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur pour %s: %v", name, err)
		}
		values[name] = value
		totalValue += value
	}

	return values, totalValue, nil
}

// PrintPortfolioSummary affiche un résumé du portefeuille
func (p *Portfolio) PrintPortfolioSummary() {
	fmt.Println("=== RÉSUMÉ DU PORTEFEUILLE ===\n")

	for name, inv := range p.Investments {
		fmt.Printf("Investissement: %s\n", name)
		fmt.Printf("  Montant investi: %.2f€\n", inv.AmountInvested)
		fmt.Printf("  Taux de référence: %.2f%%\n", inv.ReferenceRate)
		fmt.Printf("  Date d'investissement: %s\n", inv.InvestmentDate)

		if len(inv.NAVHistory) > 0 {
			latestNAV, _ := inv.GetLatestNAV()
			fmt.Printf("  Dernière NAV: %.2f€ (date: %s)\n", latestNAV.Value, latestNAV.Date)

			if len(inv.NAVHistory) >= 2 {
				performanceRate, _ := inv.CalculatePerformanceRate()
				fmt.Printf("  Taux de performance annuel: %.2f%%\n", performanceRate)
			}
		} else {
			fmt.Println("  Aucune NAV enregistrée")
		}
		fmt.Println()
	}
}

func main() {
	// Créer un portefeuille
	portfolio := NewPortfolio()

	// Ajouter des investissements
	portfolio.AddInvestment("Action Tech", 5000, 8.0, "2024-01-01")
	portfolio.AddInvestment("Obligation Corp", 3000, 4.5, "2024-01-01")
	portfolio.AddInvestment("Fonds Immobilier", 4000, 6.0, "2024-01-01")

	// Ajouter les NAV historiques
	// Action Tech
	portfolio.AddNAV("Action Tech", "2024-01-01", 5000)
	portfolio.AddNAV("Action Tech", "2024-07-01", 5300)
	portfolio.AddNAV("Action Tech", "2026-01-15", 6200)

	// Obligation Corp
	portfolio.AddNAV("Obligation Corp", "2024-01-01", 3000)
	portfolio.AddNAV("Obligation Corp", "2024-07-01", 3067)
	portfolio.AddNAV("Obligation Corp", "2026-01-15", 3235)

	// Fonds Immobilier
	portfolio.AddNAV("Fonds Immobilier", "2024-01-01", 4000)
	portfolio.AddNAV("Fonds Immobilier", "2024-07-01", 4150)
	portfolio.AddNAV("Fonds Immobilier", "2026-01-15", 4650)

	// Afficher le résumé
	portfolio.PrintPortfolioSummary()

	// Projeter la valeur du portefeuille à une date future
	projectionDate := "2027-01-15"
	fmt.Printf("=== PROJECTION AU %s ===\n\n", projectionDate)

	values, totalValue, err := portfolio.GetPortfolioValue(projectionDate)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return
	}

	for name, value := range values {
		fmt.Printf("%s: %.2f€\n", name, value)
	}

	fmt.Printf("\nValeur totale du portefeuille: %.2f€\n", totalValue)

	// Valeur initiale totale
	totalInvested := 0.0
	for _, inv := range portfolio.Investments {
		totalInvested += inv.AmountInvested
	}

	gain := totalValue - totalInvested
	gainPercent := (gain / totalInvested) * 100
	fmt.Printf("Montant investi total: %.2f€\n", totalInvested)
	fmt.Printf("Gain/Perte: %.2f€ (%.2f%%)\n", gain, gainPercent)
}
