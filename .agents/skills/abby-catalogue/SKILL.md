---
name: abby-catalogue
description: >-
  Gestion du catalogue de produits et services dans Abby via le CLI (/v2/catalog,
  /v2/catalog/product) : recherche, création de produits ou prestations avec prix
  en centimes/euros, unités, codes TVA et mise à jour. Utiliser ce skill pour toute
  demande liée aux produits ou au catalogue.
---

# Skill : Catalogue de Produits & Services Abby (`abby-catalogue`)

Ce skill permet de piloter le catalogue de produits et services d'Abby (`https://docs.abby.fr/api/produits`) via la commande `abby catalog` (alias `abby products` ou `abby produits`).

## Commandes CLI (`abby catalog`)

### 1. Lister et rechercher des produits (`GET /v2/catalog`)
Avant de créer un produit ou d'ajouter une ligne récurrente sur une facture/devis, cherchez s'il existe déjà dans le catalogue :
```powershell
# Lister les produits
./bin/abby.exe catalog list -o json

# Rechercher un produit par désignation ou référence
./bin/abby.exe catalog list --search "Développement" -o json

# Filtrer les prestations compatibles Services à la Personne (SAP)
./bin/abby.exe catalog list --sap -o json
```

### 2. Consulter un produit par ID (`GET /v2/catalog/product/{productId}`)
```powershell
./bin/abby.exe catalog get <productId> -o json
```

### 3. Créer un produit ou service (`POST /v2/catalog/product`)
Les prix sont exprimés en **centimes** (`--price-cents 35000` = `350,00 €`) ou convertis depuis des euros (`--price-euros 350`).
```powershell
./bin/abby.exe catalog create `
  --designation "TJM Développement Go" `
  --description "Journée d'ingénierie logicielle Backend Go" `
  --reference "DEV-GO-DAY" `
  --type "service_delivery" `
  --unit "day" `
  --price-cents 35000 `
  --vat "FR_2000" `
  -o json
```

### 4. Mettre à jour un produit (`PUT /v2/catalog/product/{productId}`)
```powershell
./bin/abby.exe catalog update <productId> `
  --designation "TJM Développement Go Senior" `
  --reference "DEV-GO-DAY" `
  --type "service_delivery" `
  --unit "day" `
  --price-cents 38000 `
  --vat "FR_2000" `
  -o json
```

---

## Valeurs de référence du Catalogue Abby

### Types de produits (`--type`)
| Valeur | Description |
| :--- | :--- |
| `service_delivery` | Prestation de services (BNC / BIC) |
| `sale_of_goods` | Vente de marchandises |
| `commercial_or_craft_services` | Prestations commerciales ou artisanales |
| `sale_of_manufactured_goods` | Vente de produits fabriqués |
| `disbursement` | Débours (toujours 0% TVA) |

### Unités courantes (`--unit`)
`unit`, `hour`, `day`, `week`, `fixed_rate` (forfait), `month`, `year`, `kilometer`, `kilogram`, `liter`, `batch`, `meter`, `square_meter`, `page`, `word`, `line`, `license`, `article`.

### Codes TVA (`--vat`)
`FR_2000` (20%), `FR_1000` (10%), `FR_850` (8,5%), `FR_550` (5,5%), `FR_210` (2,1%), `FR_00HT` (0% Exonéré / Franchise en base), `FR_00UE` (0% UE), `FR_0HUE` (0% Hors UE).
