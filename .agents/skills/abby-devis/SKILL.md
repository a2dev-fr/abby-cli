---
name: abby-devis
description: >-
  Gestion des devis et bons de commande Abby via le CLI : création, lignes, dates
  d'expiration, finalisation, signature/refus, création de facture d'acompte et
  création de facture finale de solde. Utiliser ce skill pour tout flux lié aux devis.
---

# Skill : Gestion des Devis & Acomptes Abby (`abby-devis`)

Ce skill décrit le pilotage des devis (`estimate`) et bons de commande (`purchase_order`) via l'API Abby v2 (`https://docs.abby.fr/api/devis`).

## Flux de travail : Du Devis à la Facture d'Acompte et Facture Finale

```text
1. Créer le devis en brouillon + lignes (estimates create / set-lines)
   ↓
2. Finaliser le devis (estimates finalize) → Attribue un numéro DE-YYYY-XXXX
   ↓
3. Marquer le devis comme signé après accord client (estimates sign)
   ↓
4. (Optionnel) Créer une facture d'acompte (estimates create-advance --percentage 30)
   → Finaliser l'acompte (billings finalize <advanceId>)
   ↓
5. Créer la facture finale de solde (estimates create-final-invoice <estimateId>)
   → Les acomptes finalisés/payés sont automatiquement déduits !
   → Finaliser la facture finale (invoices finalize <invoiceId>)
```

---

## Commandes détaillées (`abby estimates`)

### 1. Lister et consulter les devis
```powershell
./bin/abby.exe estimates list -o json
./bin/abby.exe estimates list --state signed -o json
./bin/abby.exe estimates get <estimateId> -o json
```

### 2. Créer un devis (ou bon de commande) avec ses lignes
```powershell
./bin/abby.exe estimates create <customerId> `
  --title "Refonte Site E-Commerce" `
  --designation "Conception & Développement" `
  --unit-price-cents 200000 `
  --quantity 1 `
  --unit "fixed_rate" `
  --type "service_delivery" `
  --vat "FR_2000" `
  -o json
```
*(Ajoutez `--purchase-order` pour créer un bon de commande au lieu d'un devis).*

### 3. Définir la date d'expiration (`update-timeline`)
```powershell
./bin/abby.exe estimates update-timeline <estimateId> `
  --emitted-at "2026-11-12" `
  --expired-at "2026-12-12" `
  --payment-delay "thirty_days" `
  -o json
```

### 4. Finaliser, Signer ou Refuser un devis
```powershell
# Finaliser le devis (après confirmation utilisateur)
./bin/abby.exe estimates finalize <estimateId> -o json

# Marquer comme signé par le client
./bin/abby.exe estimates sign <estimateId> -o json

# Annuler la signature si besoin
./bin/abby.exe estimates unsign <estimateId> -o json

# Marquer comme refusé
./bin/abby.exe estimates refuse <estimateId> -o json
```

### 5. Générer une facture d'acompte depuis un devis signé (`create-advance`)
Le devis **doit être signé** (`state: "signed"`).
```powershell
# Acompte en pourcentage (ex: 30%)
./bin/abby.exe estimates create-advance <estimateId> --percentage 30 -o json

# Ou acompte d'un montant fixe HT en centimes (ex: 350,00 € HT = 35000 centimes)
./bin/abby.exe estimates create-advance <estimateId> --amount-cents 35000 -o json
```
L'acompte est créé en brouillon (`type: "advance"`). Pour le finaliser :
```powershell
./bin/abby.exe billings finalize <advanceId> -o json
```

### 6. Générer la facture finale de solde (`create-final-invoice`)
Depuis un devis signé, génère la facture finale en déduisant automatiquement les acomptes :
```powershell
./bin/abby.exe estimates create-final-invoice <estimateId> -o json
```
Vous pouvez ensuite finaliser et encaisser cette facture via `abby invoices finalize <invoiceId>` et `abby invoices mark-paid <invoiceId>`.
