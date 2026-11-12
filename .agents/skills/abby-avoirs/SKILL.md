---
name: abby-avoirs
description: >-
  Gestion des avoirs (notes de crédit / assets) dans Abby via le CLI : création
  d'un avoir depuis une facture ou un acompte, création d'un avoir vierge, édition
  des lignes, finalisation et téléchargement PDF.
---

# Skill : Gestion des Avoirs Abby (`abby-avoirs`)

Ce skill décrit la création et la gestion des avoirs (`asset`) sur l'API Abby v2 (`https://docs.abby.fr/api/avoirs`).

## Quand utiliser un Avoir vs Annuler une Facture ?

1. **Annulation totale et immédiate d'une facture finalisée** :
   Utilisez directement `abby invoices cancel <invoiceId> -o json`. Cela crée **et finalise** automatiquement l'avoir correspondant à 100% de la facture.
2. **Remboursement partiel ou avoir personnalisé (en brouillon)** :
   Utilisez `abby assets create-from-invoice <invoiceId>` (ou `create <customerId>` pour un avoir vierge), modifiez les lignes avec `abby assets set-lines <assetId>`, puis finalisez avec `abby assets finalize <assetId>`.

---

## Commandes détaillées (`abby assets`)

### 1. Lister et consulter les avoirs
```powershell
./bin/abby.exe assets list -o json
./bin/abby.exe assets get <assetId> -o json
```

### 2. Créer un avoir en brouillon depuis une facture ou un acompte
```powershell
# Depuis une facture finalisée : reprend les lignes de la facture en brouillon d'avoir
./bin/abby.exe assets create-from-invoice <invoiceId> -o json

# Depuis une facture d'acompte :
./bin/abby.exe assets create-from-advance <advanceId> -o json
```

### 3. Créer un avoir vierge pour un client (`POST /v2/billing/asset/{customerId}`)
```powershell
./bin/abby.exe assets create <customerId> `
  --designation "Geste commercial" `
  --unit-price-cents 10000 `
  --quantity 1 `
  --unit "unit" `
  --type "service_delivery" `
  --vat "FR_2000" `
  -o json
```

### 4. Modifier les lignes d'un avoir brouillon (`set-lines`)
```powershell
./bin/abby.exe assets set-lines <assetId> `
  --designation "Remboursement partiel prestation" `
  --unit-price-cents 25000 `
  --quantity 1 `
  --vat "FR_2000" `
  -o json
```

### 5. Finaliser et télécharger l'avoir
Demandez toujours confirmation à l'utilisateur avant de finaliser l'avoir :
```powershell
./bin/abby.exe assets finalize <assetId> -o json
./bin/abby.exe assets download <assetId> --out "avoir.pdf"
```
