---
name: abby-cli
description: >-
  Skill maître d'orchestration pour piloter 100% de la plateforme Abby via le CLI Go
  (Factures, Devis, Acomptes, Avoirs, Clients Particuliers & Entreprises, Catalogue Produits,
  Livre des recettes, Livre des achats, Fournisseurs, Plan comptable, CRM Opportunités et
  endpoints avancés via `abby raw`). Activer ce skill pour toute demande complexe ou
  multi-domaine sur Abby.
---

# Skill Maître : Orchestration Complète Abby CLI (`abby-cli`)

Ce skill permet d'exploiter **100 % du potentiel** du binaire Go `abby` et de l'API REST v2/v3 d'Abby (`https://docs.abby.fr`).

---

## 1. Prérequis & Auto-Vérification du Binaire

Avant d'exécuter une commande, vérifiez que le binaire existe. S'il est absent, compilez-le immédiatement :
```powershell
if (-not (Test-Path "./bin/abby.exe")) {
  go build -o bin/abby.exe ./cmd/abby
}
```
Utilisez **toujours** le flag `-o json` (`--output json`) lorsque vous pilotez le CLI afin de récupérer des objets JSON structurés (`id`, `number`, `state`, `finalizable`, `finalizeRequirements`, `total`).

---

## 2. Cartographie Complète des 12 Groupes de Commandes

| Commande CLI | Alias | Domaine couvert | Skill détaillé |
| :--- | :--- | :--- | :--- |
| `abby auth` | - | Clé API (`set-key`, `status`, `whoami`, `clear`) | `abby-auth` |
| `abby contacts` | `contact` | Clients particuliers (`list`, `get`, `create`, `update`, `delete`) | `abby-clients` |
| `abby orgs` | `org`, `organizations` | Entreprises clientes (`list`, `get`, `create`, `update`, `delete`, `contacts`, `add-contact`, `set-default-contact`) | `abby-clients` |
| `abby invoices` | `invoice`, `factures` | Factures (`list`, `get`, `create`, `set-lines`, `update-info`, `update-timeline`, `set-locale`, `set-currency`, `finalize`, `activate-online-payment`, `mark-paid`, `mark-unpaid`, `cancel`, `download`, `send-email`) | `abby-factures` |
| `abby estimates` | `estimate`, `devis` | Devis & Bons de commande (`list`, `get`, `create`, `set-lines`, `update-info`, `update-timeline`, `finalize`, `sign`, `unsign`, `refuse`, `unrefuse`, `create-advance`, `create-final-invoice`, `download`, `send-email`) | `abby-devis` |
| `abby advances` | `advance`, `acomptes` | Factures d'acompte (`list`, `get`, `update-amount`, `finalize`, `mark-paid`, `cancel`, `download`) | `abby-devis` |
| `abby assets` | `asset`, `avoirs` | Avoirs (`list`, `get`, `create`, `create-from-invoice`, `create-from-advance`, `set-lines`, `update-info`, `update-timeline`, `finalize`, `download`, `send-email`) | `abby-avoirs` |
| `abby billings` | `billing` | Opérations transverses sur tout document (`list`, `get`, `set-lines`, `set-title`, `finalize`, `download`, `send-email`, `archive`, `unarchive`, `delete`) | `abby-factures` |
| `abby catalog` | `products`, `produits` | Catalogue produits & services (`list`, `get`, `create`, `update`) | `abby-catalogue` |
| `abby accounting` | `compta`, `books` | Comptabilité : `incomes` (Recettes), `purchases` (Achats v3/v2), `providers` (Fournisseurs), `accounts` (Plan comptable), `transaction` | `abby-comptabilite` |
| `abby opportunities` | `opportunity`, `crm` | Pipeline commercial CRM (`categories`, `get`, `create`, `update`, `delete`) | `abby-devis` |
| `abby raw` | - | Appel HTTP générique authentifié (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) sur n'importe quel endpoint Abby | Voir Section 4 |

---

## 3. Flux de Travail Multi-Domaines (Chaînage Agent)

### Flux A : Cycle Commercial Complet (Client → Produit → Devis → Acompte 30% → Facture de Solde)
1. **Trouver ou créer le client** :
   ```powershell
   ./bin/abby.exe orgs list --search "Martin SAS" -o json
   # Si absent : ./bin/abby.exe orgs create --name "Martin SAS" --siret "..." --address "..." --zip "..." --city "..." -o json
   ```
2. **Créer le devis en brouillon avec lignes et remises éventuelles** :
   ```powershell
   ./bin/abby.exe estimates create <customerId> --title "Refonte E-Commerce" --designation "Développement" --unit-price-cents 200000 --quantity 1 --unit fixed_rate --vat FR_2000 -o json
   ```
3. **Finaliser et signer le devis** *(après confirmation utilisateur)* :
   ```powershell
   ./bin/abby.exe estimates finalize <estimateId> -o json
   ./bin/abby.exe estimates sign <estimateId> -o json
   ```
4. **Créer, finaliser et encaisser l'acompte de 30%** :
   ```powershell
   ./bin/abby.exe estimates create-advance <estimateId> --percentage 30 -o json
   ./bin/abby.exe advances finalize <advanceId> -o json
   ./bin/abby.exe advances mark-paid <advanceId> --method transfer -o json
   ```
5. **Générer la facture finale de solde** (déduit automatiquement l'acompte payé) :
   ```powershell
   ./bin/abby.exe estimates create-final-invoice <estimateId> -o json
   ./bin/abby.exe invoices finalize <finalInvoiceId> -o json
   ```

---

## 4. Mode Universel `abby raw` : Exploiter 100% des Endpoints Avancés

Pour tous les endpoints spécialisés de l'API Abby qui ne nécessitent pas de sous-commande dédiée, utilisez **`abby raw <METHOD> <PATH> [-d '<JSON>'] [-q clé=valeur]`** :

### A. Factures récurrentes / Abonnements (`/v2/billing/invoice/{invoiceId}/frequency`)
```powershell
# Créer ou modifier une récurrence sur une facture brouillon
./bin/abby.exe raw POST /v2/billing/invoice/<invoiceId>/frequency -d '{
  "frequency": "MONTHLY",
  "startAt": 1767225600,
  "endAt": 1798761600,
  "autoFinalize": false,
  "autoSendByEmail": false
}'

# Supprimer la récurrence d'une facture
./bin/abby.exe raw DELETE /v2/billing/invoice/<invoiceId>/frequency
```

### B. Configurer les lignes d'acompte prévisionnelles sur un devis (`/v2/billing/estimate/{estimateId}/advance-lines`)
```powershell
./bin/abby.exe raw PATCH /v2/billing/estimate/<estimateId>/advance-lines -d '{
  "advanceLines": [
    {"mode": "PERCENTAGE", "percentage": 30},
    {"mode": "PERCENTAGE", "percentage": 40}
  ]
}'
```

### C. Personnalisation avancée d'un document (`bank-information`, `delivery-address`, `display-settings`, `theme`)
```powershell
# Modifier l'adresse de livraison d'un document
./bin/abby.exe raw PATCH /v2/billing/<billingId>/delivery-address -d '{
  "displayDeliveryAddress": true,
  "deliveryAddress": {
    "address": "15 Rue du Port",
    "zipCode": "44000",
    "city": "Nantes",
    "country": "FR"
  }
}'

# Modifier les coordonnées bancaires affichées sur un document
./bin/abby.exe raw PATCH /v2/billing/<billingId>/bank-information -d '{
  "iban": "FR7630003500011234567890189",
  "bic": "AGRIFRPP",
  "holderName": "Ma Société SAS"
}'

# Activer/désactiver les colonnes ou paramètres d'affichage
./bin/abby.exe raw PATCH /v2/billing/<billingId>/display-settings -d '{
  "displayUnit": true,
  "displayReference": true,
  "displayDescription": true
}'

# Lister les thèmes de facturation et appliquer un thème à un document
./bin/abby.exe raw GET /v2/billing/theme
./bin/abby.exe raw PATCH /v2/billing/<billingId>/theme -d '{"themeId": "<themeId>"}'
```

### D. Prévisualiser et tester un email de facturation avant envoi
```powershell
# Rendu HTML de l'email (emailType: invoice, estimate, asset, advance, reminder)
./bin/abby.exe raw GET /v2/billing/<billingId>/email/invoice/render

# Envoyer un email de test à soi-même
./bin/abby.exe raw POST /v2/billing/<billingId>/email/invoice/test
```

### E. Portail Client, Paiements en ligne, Logs d'audit & Webhooks
```powershell
# Consulter le compte de paiement en ligne et les paiements du portail client
./bin/abby.exe raw GET /v2/billing/payment-account
./bin/abby.exe raw GET /v2/customer-portal/payments

# Consulter l'historique / dernier log d'une ressource
./bin/abby.exe raw GET /logs/<resourceId>

# Créer ou supprimer un Webhook entreprise
./bin/abby.exe raw POST /webhook -d '{"url": "https://example.com/webhook", "events": ["invoice.finalized", "invoice.paid"]}'
./bin/abby.exe raw DELETE /webhook/<webhookId>
```
