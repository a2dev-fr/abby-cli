# Abby CLI (`abby-cli`) — Client Go & Skills IA pour l'API REST Abby

`abby-cli` est une application en ligne de commande écrite en **Go** permettant de piloter l'**API REST v2 d'Abby** ([docs.abby.fr/api/factures](https://docs.abby.fr/api/factures)), accompagnée d'un jeu de **Skills et Règles modulaires** (`.agents/`) permettant à un agent IA de gérer de bout en bout les clients, devis, acomptes, factures et avoirs.

---

## Architecture du Projet

```text
c:\code\abby-cli\
├── cmd/
│   └── abby/
│       └── main.go                  # Point d'entrée minimaliste du binaire CLI
├── internal/
│   ├── config/
│   │   └── config.go                # Résolution config (flag > env ABBY_API_KEY > ~/.config/abby-cli/config.json)
│   ├── api/
│   │   ├── client.go                # Client HTTP générique (Bearer suk-***, HTTPS strict, gestion body nil)
│   │   ├── errors.go                # Décodage structuré des erreurs API Abby (api_error, authentication_error, rsc_error)
│   │   ├── models.go                # DTOs et énumérations v2 (centimes, codes TVA FR_2000, timestamps secondes)
│   │   ├── company.go               # Service Company (/v2/company/me)
│   │   ├── contacts.go              # Service Clients Particuliers (/contacts, /contact/{id})
│   │   ├── organizations.go         # Service Clients Professionnels (/organizations, /organization/{id})
│   │   ├── billings.go              # Service Billing transverse (/v2/billings, lignes, finalize, PDF, email, archive)
│   │   ├── invoices.go              # Service Factures (/v2/billing/invoice/*, reconciliate/paid, mark-as-unpaid, cancel)
│   │   ├── estimates.go             # Service Devis & Acomptes (/v2/billing/estimate/*, /v2/billing/advance/*)
│   │   ├── assets.go                # Service Avoirs (/v2/billing/asset/*)
│   │   ├── catalog.go               # Service Catalogue Produits (/v2/catalog, /v2/catalog/product/*)
│   │   ├── accounting.go            # Service Comptabilité (/incomeBook, /v3/purchaseRegister, /providers, /v2/accounting-accounts)
│   │   └── opportunities.go         # Service CRM / Opportunités (/v2/opportunity/*)
│   ├── output/
│   │   └── printer.go               # Affichage double mode : Table (humain) ou JSON (agent IA), conversions € <-> centimes
│   └── cli/
│       ├── root.go                  # Commande racine Cobra + flags persistants (--api-key, --base-url, -o json|table)
│       ├── auth.go                  # Sous-commandes `abby auth`
│       ├── contacts.go              # Sous-commandes `abby contacts`
│       ├── organizations.go         # Sous-commandes `abby orgs`
│       ├── invoices.go              # Sous-commandes `abby invoices`
│       ├── estimates.go             # Sous-commandes `abby estimates`
│       ├── advances.go              # Sous-commandes `abby advances`
│       ├── assets.go                # Sous-commandes `abby assets`
│       ├── billings.go              # Sous-commandes `abby billings`
│       ├── catalog.go               # Sous-commandes `abby catalog`
│       ├── accounting.go            # Sous-commandes `abby accounting` (incomes, purchases, providers, accounts, transaction)
│       ├── opportunities.go         # Sous-commandes `abby opportunities`
│       └── raw.go                   # Sous-commande `abby raw <METHOD> <PATH>`
└── .agents/
    ├── rules/
    │   └── abby-conventions.md      # Règles permanentes (centimes, timestamps secondes, confirmation avant finalize)
    └── skills/
        ├── abby-auth/SKILL.md         # Skill Authentification & Diagnostic (/v2/company/me)
        ├── abby-clients/SKILL.md      # Skill Clients Particuliers (Contacts) & Entreprises (Organizations)
        ├── abby-factures/SKILL.md     # Skill Cycle de vie des Factures (brouillon -> lignes -> finalisation -> paiement)
        ├── abby-devis/SKILL.md        # Skill Devis, Signatures, Factures d'acompte, Factures de solde & CRM
        ├── abby-avoirs/SKILL.md       # Skill Avoirs & Annulations de factures
        ├── abby-catalogue/SKILL.md    # Skill Catalogue de produits et services (/v2/catalog)
        └── abby-comptabilite/SKILL.md # Skill Comptabilité (Livre des recettes, Livre des achats, Fournisseurs, Plan comptable)
```

---

## Installation & Configuration

### 1. Compiler le binaire
```powershell
go build -o bin/abby.exe ./cmd/abby
```

### 2. Configurer la clé API Abby (`Bearer suk-*****`)
Générez une clé API depuis votre compte Abby dans **Paramètres > Intégrations** (`https://my.app-abby.com/settings/integrations`), puis enregistrez-la :

```powershell
# Via la commande dédiée (stockée dans %AppData%\abby-cli\config.json avec permissions restreintes)
./bin/abby.exe auth set-key "suk-votre_cle_api"

# Ou via variable d'environnement
$env:ABBY_API_KEY = "suk-votre_cle_api"

# Vérifier la connexion au compte Abby
./bin/abby.exe auth whoami
```

---

## Exemples d'utilisation CLI

### Clients (`contacts` & `orgs`)
```powershell
# Rechercher une entreprise cliente
./bin/abby.exe orgs list --search "Dupont"

# Créer une entreprise cliente
./bin/abby.exe orgs create --name "Dupont SARL" --siret "12345678900012" --email "contact@dupont.fr" --address "10 Rue de la Paix" --zip "75002" --city "Paris"
```

### Factures (`invoices`)
```powershell
# Créer une facture en brouillon avec une ligne à 350,00 € HT (35000 centimes) et TVA 20%
./bin/abby.exe invoices create <customerId> `
  --title "Prestation Mai 2026" `
  --designation "Développement Backend Go" `
  --unit-price-cents 35000 `
  --quantity 2 `
  --unit "day" `
  --vat "FR_2000"

# Finaliser la facture (attribue le numéro officiel et verrouille la facture)
./bin/abby.exe invoices finalize <invoiceId>

# Télécharger le PDF
./bin/abby.exe invoices download <invoiceId> --out facture.pdf

# Marquer la facture comme payée (encaisse automatiquement le solde TTC restant)
./bin/abby.exe invoices mark-paid <invoiceId> --method transfer
```

### Devis & Acomptes (`estimates`)
```powershell
# Créer un devis en brouillon
./bin/abby.exe estimates create <customerId> --designation "Refonte Site Web" --unit-price-euros 2000 --vat "FR_2000"

# Finaliser puis signer le devis
./bin/abby.exe estimates finalize <estimateId>
./bin/abby.exe estimates sign <estimateId>

# Générer un acompte de 30% puis la facture finale de solde
./bin/abby.exe estimates create-advance <estimateId> --percentage 30
./bin/abby.exe estimates create-final-invoice <estimateId>
```

### Avoirs (`assets`)
```powershell
# Annuler intégralement une facture finalisée (crée et finalise un avoir automatiquement)
./bin/abby.exe invoices cancel <invoiceId>

# Ou créer un avoir en brouillon à partir d'une facture pour l'éditer avant finalisation
./bin/abby.exe assets create-from-invoice <invoiceId>
```

### Catalogue Produits (`catalog`)
```powershell
# Lister ou chercher des produits
./bin/abby.exe catalog list --search "Développement"

# Créer un produit / prestation dans le catalogue (350,00 € HT = 35000 centimes)
./bin/abby.exe catalog create --designation "TJM Développement Go" --type service_delivery --unit day --price-cents 35000 --vat FR_2000
```

### Comptabilité : Livre des Recettes, Livre des Achats & Fournisseurs (`accounting`)
```powershell
# Ajouter une entrée au Livre des recettes (/incomeBook)
./bin/abby.exe accounting incomes create --client "Dupont SARL" --ht-cents 100000 --ttc-cents 120000 --vat-cents 20000 --vat-id 1 --product-type 3

# Créer un fournisseur puis enregistrer une dépense dans le Livre des achats (/v3/purchaseRegister)
./bin/abby.exe accounting providers create --name "OVHcloud SAS" --siret "42476141900045"
./bin/abby.exe accounting purchases create --provider-id <providerId> --label "Serveur Dédié" --amount-cents 5999 --vat-id 1 --account-number 606300

# Consulter le plan comptable (/v2/accounting-accounts)
./bin/abby.exe accounting accounts
```

### Requêtes HTTP brutes (`raw`)
```powershell
# Appeler n'importe quel endpoint de l'API Abby avec authentification Bearer automatique
./bin/abby.exe raw GET /v2/company/me
./bin/abby.exe raw GET /v2/billings -q page=1 -q limit=5 -q test=false
```
