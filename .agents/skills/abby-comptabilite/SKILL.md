---
name: abby-comptabilite
description: >-
  Gestion de la comptabilité dans Abby via le CLI : Livre des recettes (/incomeBook),
  Livre des achats (/v3/purchaseRegister et /v2/purchaseRegister), Fournisseurs (/providers),
  Plan comptable (/v2/accounting-accounts) et Transactions bancaires (/v2/transactions).
  Utiliser ce skill pour toute demande liée aux recettes, achats, dépenses, fournisseurs ou comptabilité.
---

# Skill : Comptabilité, Livre des Recettes & Livre des Achats Abby (`abby-comptabilite`)

Ce skill permet de piloter les livres comptables et les fournisseurs d'Abby (`https://docs.abby.fr/api/livre-des-recettes` et `https://docs.abby.fr/api/livre-des-achats`) via la commande `abby accounting` (alias `abby compta`).

---

## 1. Livre des Recettes (`abby accounting incomes`)

Permet d'enregistrer ou de supprimer une recette manuelle dans le **Livre des recettes** (`/incomeBook`).

### Ajouter une recette (`POST /incomeBook`)
```powershell
./bin/abby.exe accounting incomes create `
  --client "Dupont SARL" `
  --reference "REC-2026-01" `
  --paid-at "2026-11-12" `
  --ht-cents 100000 `
  --vat-cents 20000 `
  --ttc-cents 120000 `
  --vat-id 1 `
  --product-type 3 `
  --payment-method 1 `
  -o json
```
- `--product-type` : `1` = Vente de marchandises, `2` = Prestation de services (BIC), `3` = Prestation de services (BNC), `4` = Produits fabriqués, `5` = Débours.
- `--vat-id` : `1` = 20%, `2` = 10%, `3` = 5,5%, `4` = 2,1%, `5` = 0% (Franchise HT), `6` = 8,5%.
- `--payment-method` : `1` = Virement, `2` = Carte bancaire, `3` = Espèces, `4` = Chèque, `5` = Prélèvement.

### Supprimer une recette (`DELETE /incomeBook/{id}`)
```powershell
./bin/abby.exe accounting incomes delete <incomeId> -o json
```

---

## 2. Fournisseurs (`abby accounting providers`)

**Règle critique** : Pour créer une dépense dans le **Livre des achats**, l'identifiant `--provider-id` (`thirdPartyId`) doit obligatoirement être un **Fournisseur** (`PROVIDER`), et **non** un contact ou une organisation cliente.

### Lister ou chercher un fournisseur (`GET /providers`)
```powershell
./bin/abby.exe accounting providers list --search "OVH" -o json
```

### Créer ou mettre à jour un fournisseur (`POST /provider`, `PUT /provider/{id}`)
```powershell
./bin/abby.exe accounting providers create `
  --name "OVHcloud SAS" `
  --siret "42476141900045" `
  --vat-number "FR22424761419" `
  --address "2 Rue Kellermann" `
  --zip "59100" `
  --city "Roubaix" `
  --country "FR" `
  -o json

./bin/abby.exe accounting providers update <providerId> --name "OVHcloud SAS" --siret "42476141900045" -o json
```

### Supprimer un fournisseur (`DELETE /provider/{id}`)
```powershell
./bin/abby.exe accounting providers delete <providerId> -o json
```

---

## 3. Livre des Achats (`abby accounting purchases`)

Gère les dépenses fournisseurs sur `/v3/purchaseRegister` (et `/v2/purchaseRegister` via `--v2`).

### Flux de travail pour enregistrer un achat
1. Chercher le fournisseur avec `abby accounting providers list --search "<nom>" -o json` (ou le créer s'il n'existe pas).
2. Optionnel : Vérifier le numéro de compte comptable avec `abby accounting accounts -o json` (ex: `606300` Fournitures, `622600` Honoraires).
3. Créer l'écriture d'achat avec `abby accounting purchases create`.

### Lister et consulter les achats (`GET /v3/purchaseRegister/list`)
```powershell
./bin/abby.exe accounting purchases list -o json
./bin/abby.exe accounting purchases list --from 2026-01-01 --to 2026-12-31 --search "Serveur" -o json
./bin/abby.exe accounting purchases get <purchaseId> -o json
```

### Créer un achat (`POST /v3/purchaseRegister`)
```powershell
./bin/abby.exe accounting purchases create `
  --provider-id <providerId> `
  --label "Hébergement Serveur Dédié Novembre" `
  --reference "FR-OVH-99812" `
  --date "2026-11-12" `
  --amount-cents 5999 `
  --vat-id 1 `
  --account-number 606300 `
  --payment-method "BANK_CARD" `
  -o json
```
- `--amount-cents` (ou `--amount-euros`) : Montant **TTC** en centimes (utilisez un montant négatif pour un remboursement).
- `--vat-id` : `1` = 20%, `2` = 10%, `3` = 5,5%, `4` = 2,1%, `5` = 0% HT, `6` = 8,5% DOM, `7` = UE (autoliquidation), `8` = Hors UE.

### Ventilation multi-taux TVA ou multi-comptes (`--entries-json`) et mise à jour (`update`)
Lorsqu'une même facture d'achat comporte plusieurs taux de TVA (ex: restauration 10% + 20%) ou une part personnelle (`isPersonal: true`), ventilez-la avec `--entries-json` :
```powershell
./bin/abby.exe accounting purchases create `
  --provider-id <providerId> `
  --label "Note de frais Restaurant & Déplacement" `
  --date "2026-11-12" `
  --entries-json '[
    {"amount": 3300, "isPersonal": false, "vatId": 2, "accountingAccountNumber": 625600},
    {"amount": 1200, "isPersonal": false, "vatId": 1, "accountingAccountNumber": 625100}
  ]' `
  -o json

# Mettre à jour un achat existant (PUT /v3/purchaseRegister/{id})
./bin/abby.exe accounting purchases update <purchaseId> --provider-id <providerId> --label "Libellé corrigé" --amount-cents 4500 --vat-id 1 -o json
```

### Supprimer un achat (`DELETE /v3/purchaseRegister/{id}`)
```powershell
./bin/abby.exe accounting purchases delete <purchaseId> -o json
```

---

## 4. Plan Comptable & Transactions Bancaires

```powershell
# Lister le plan comptable de l'entreprise (GET /v2/accounting-accounts)
./bin/abby.exe accounting accounts -o json

# Inspecter une transaction bancaire (GET /v2/transactions/{id})
./bin/abby.exe accounting transaction <transactionId> -o json
```
