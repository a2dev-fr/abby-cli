---
description: Conventions obligatoires pour l'utilisation du CLI Go Abby et de l'API REST Abby v2
trigger: always_on
---

# Règles et Conventions Abby CLI & API REST

Lorsque vous pilotez la plateforme Abby via le binaire `abby` (`./bin/abby.exe` ou `go run ./cmd/abby`) ou que vous modifiez ce dépôt :

1. **Format de sortie pour l'agent (`--output json`)** :
   - Utilisez systématiquement le flag `-o json` (ou `--output json`) lorsque vous exécutez `abby` pour inspecter des identifiants (`id`), vérifier `finalizable`, ou enchaîner des étapes.
2. **Montants monétaires en centimes** :
   - L'API REST Abby exprime **tous** les montants monétaires en **centimes d'euros** (`10000` = `100,00 €`, `7550` = `75,50 €`).
   - Dans le CLI, privilégiez `--unit-price-cents` / `--amount-cents` pour une précision exacte, ou `--unit-price-euros` / `--amount-euros` qui convertit automatiquement en centimes.
3. **Dates en timestamps Unix (secondes)** :
   - L'API Abby attend des timestamps Unix en **secondes** (jamais en millisecondes) sur les endpoints `/timeline`.
   - Le CLI accepte directement le format `YYYY-MM-DD` (ex: `2026-11-12`), `today`, ou un timestamp en secondes.
4. **Codes TVA normalisés (API v2)** :
   - Utilisez les codes TVA officiels de l'API v2 :
     - `FR_2000` : 20% (Taux normal)
     - `FR_1000` : 10% (Taux intermédiaire)
     - `FR_850` : 8,5% (Taux DOM)
     - `FR_550` : 5,5% (Taux réduit)
     - `FR_210` : 2,1% (Taux super-réduit)
     - `FR_00HT` : 0% (Exonéré / Franchise en base de TVA)
     - `FR_00UE` : 0% (TVA Intracommunautaire UE)
     - `FR_0HUE` : 0% (Hors UE)
5. **Irréversibilité de la finalisation (Loi anti-fraude TVA)** :
   - Tout document créé (`invoice`, `estimate`, `asset`, `advance`) démarre à l'état `draft` (brouillon).
   - La commande `finalize` (`PATCH /v2/billing/{billingId}/finalize`) attribue un numéro officiel et **verrouille définitivement** le document.
   - **Demandez toujours confirmation à l'utilisateur** (en lui présentant un récapitulatif : client, lignes, total HT/TTC) avant d'exécuter `finalize`, `cancel` ou `send-email`.
   - Une facture finalisée ne peut plus être modifiée ni supprimée : pour l'annuler, utilisez `abby invoices cancel <invoiceId>` ou créez un avoir (`abby assets create-from-invoice <invoiceId>`).
