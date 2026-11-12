---
name: abby-factures
description: >-
  Pilotage complet des factures Abby via le CLI (création brouillon, ajout de lignes
  en centimes/euros, dates timeline en secondes, finalisation irréversible, marquage payé/impayé,
  téléchargement PDF, envoi email et annulation). Utiliser ce skill pour toute demande liée aux factures.
---

# Skill : Gestion des Factures Abby (`abby-factures`)

Ce skill couvre le cycle de vie complet d'une facture sur l'API Abby v2 (`https://docs.abby.fr/api/factures`).

## Flux de travail standard : Créer, Finaliser et Envoyer une Facture

```text
1. Identifier le client (orgs list / contacts list)
   ↓
2. Créer la facture en brouillon + définir ses lignes (invoices create / set-lines)
   ↓
3. (Optionnel) Configurer les dates / notes / moyens de paiement (update-timeline / update-info)
   ↓
4. Vérifier que finalizable == true et présenter l'aperçu à l'utilisateur
   ↓
5. [CONFIRMATION UTILISATEUR REQUISE] Finaliser la facture (invoices finalize)
   ↓
6. Télécharger le PDF (invoices download) ou Envoyer par email (invoices send-email)
   ↓
7. Lors de l'encaissement : Marquer comme payée (invoices mark-paid)
```

---

## Commandes détaillées (`abby invoices`)

### 1. Lister et filtrer les factures
```powershell
# Lister toutes les factures en JSON
./bin/abby.exe invoices list -o json

# Filtrer par état (draft, finalized, paid), retard ou période
./bin/abby.exe invoices list --state finalized --late -o json
./bin/abby.exe invoices list --from 2026-01-01 --to 2026-12-31 --search "Dupont" -o json
```

### 2. Créer une facture (brouillon) avec une ligne immédiate
`customerId` peut être l'`id` d'un contact particulier ou d'une organisation :
```powershell
./bin/abby.exe invoices create <customerId> `
  --title "Prestation Développement Web" `
  --designation "Développement application Go" `
  --unit-price-cents 35000 `
  --quantity 2 `
  --unit "day" `
  --type "service_delivery" `
  --vat "FR_2000" `
  -o json
```
*(Astuce : `--unit-price-cents 35000` équivaut à `--unit-price-euros 350` soit 350,00 € HT).*

### 3. Définir plusieurs lignes sur une facture brouillon (`set-lines`)
Pour une facture multi-lignes, passez un tableau JSON via `--lines-json` ou `--lines-file` :
```powershell
./bin/abby.exe invoices set-lines <invoiceId> --lines-json '[
  {
    "designation": "Audit architecture Go",
    "description": "Revue du code et conception CLI",
    "unitPrice": 75000,
    "quantity": 1,
    "quantityUnit": "fixed_rate",
    "type": "service_delivery",
    "vatCode": "FR_2000"
  },
  {
    "designation": "Hébergement annuel",
    "unitPrice": 12000,
    "quantity": 12,
    "quantityUnit": "month",
    "type": "service_delivery",
    "vatCode": "FR_2000"
  }
]' -o json
```

### 4. Modifier l'échéancier (`update-timeline`) et les informations générales (`update-info`)
Les dates sont converties automatiquement en **timestamps Unix en secondes** :
```powershell
./bin/abby.exe invoices update-timeline <invoiceId> `
  --emitted-at "2026-11-12" `
  --payment-delay "thirty_days" `
  --delivery-date "2026-11-12" `
  -o json

./bin/abby.exe invoices update-info <invoiceId> `
  --header-note "Projet Réf #2026-42" `
  --footer-note "Merci pour votre confiance." `
  --payment-methods "transfer,credit_card" `
  -o json
```

### 5. Finaliser une facture (`finalize`) — ⚠️ Action irréversible
Avant de finaliser :
1. Vérifiez avec `./bin/abby.exe invoices get <invoiceId> -o json` que `"finalizable": true` et que `"finalizeRequirements"` est vide.
2. Demandez confirmation à l'utilisateur en lui indiquant le client et le montant Total HT / TTC.
```powershell
./bin/abby.exe invoices finalize <invoiceId> -o json
```
Une fois finalisée, la facture reçoit son numéro définitif (ex: `FA-2026-0042`) et ne peut plus être modifiée.

### 6. Marquer une facture comme payée (`mark-paid`) ou non payée (`mark-unpaid`)
```powershell
# Encaisser la totalité du reste à payer TTC automatiquement :
./bin/abby.exe invoices mark-paid <invoiceId> --method "transfer" --received-at "2026-11-12" -o json

# Ou en spécifiant le montant exact TTC en centimes :
./bin/abby.exe invoices mark-paid <invoiceId> --amount-cents 120000 --method "transfer" -o json

# Annuler le marquage payé :
./bin/abby.exe invoices mark-unpaid <invoiceId> -o json
```

### 7. Télécharger le PDF ou envoyer par email
```powershell
# Télécharger le PDF
./bin/abby.exe invoices download <invoiceId> --out "FA-2026-0042.pdf"

# Envoyer par email au client
./bin/abby.exe invoices send-email <invoiceId> `
  --to "client@example.com" `
  --subject "Votre facture FA-2026-0042" `
  --content "<p>Bonjour, veuillez trouver ci-joint votre facture.</p>" `
  --copy
```

### 8. Annuler une facture finalisée (`cancel`)
Une facture finalisée ne peut pas être supprimée (`DELETE` ne fonctionne que sur un brouillon). Pour l'annuler légalement, Abby génère et finalise automatiquement un avoir :
```powershell
./bin/abby.exe invoices cancel <invoiceId> -o json
```
