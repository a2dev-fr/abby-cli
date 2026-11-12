---
name: abby-clients
description: >-
  Gestion des clients dans Abby via le CLI : clients particuliers (Contacts) et
  clients professionnels / entreprises (Organizations, SIRET, TVA, contacts d'entreprise).
  Utiliser ce skill pour chercher, créer ou modifier un client avant d'émettre un devis ou une facture.
---

# Skill : Gestion des Clients Abby (`abby-clients`)

Dans Abby (`https://docs.abby.fr/api/clients`), un client (`customerId`) peut être :
- Un **particulier** (`Contact` — endpoints `/contacts` et `/contact/{id}`)
- Une **entreprise / organisation** (`Organization` — endpoints `/organizations` et `/organization/{id}`)

L'identifiant retourné (`id`) sert directement de `<customerId>` lors de la création d'une facture, d'un devis ou d'un avoir.

---

## 1. Clients Professionnels / Entreprises (`abby orgs`)

### Rechercher une entreprise existante
Avant de créer une entreprise, vérifiez toujours si elle existe déjà :
```powershell
./bin/abby.exe orgs list --search "Dupont" -o json
```

### Créer une entreprise (`POST /organization`)
Pour une entreprise française, renseignez autant que possible le SIRET et l'adresse complète (requis pour la finalisation des factures) :
```powershell
./bin/abby.exe orgs create `
  --name "Dupont SARL" `
  --commercial-name "Dupont Digital" `
  --siret "12345678900012" `
  --vat-number "FR12123456789" `
  --email "compta@dupont.fr" `
  --address "10 Rue de la Paix" `
  --zip "75002" `
  --city "Paris" `
  --country "FR" `
  -o json
```

### Consulter, modifier ou supprimer une entreprise
```powershell
# Détails d'une organisation
./bin/abby.exe orgs get <organizationId> -o json

# Mettre à jour une organisation
./bin/abby.exe orgs update <organizationId> --name "Dupont SAS" --siret "12345678900012" --address "10 Rue de la Paix" --zip "75002" --city "Paris" -o json

# Lister les contacts rattachés à l'organisation
./bin/abby.exe orgs contacts <organizationId> -o json

# Définir le contact par défaut d'une organisation
./bin/abby.exe orgs set-default-contact <organizationId> <contactId> -o json

# Supprimer une organisation
./bin/abby.exe orgs delete <organizationId>
```

---

## 2. Clients Particuliers (`abby contacts`)

### Rechercher un contact particulier
```powershell
./bin/abby.exe contacts list --search "Jean Martin" -o json
```

### Créer un contact particulier (`POST /contact`)
```powershell
./bin/abby.exe contacts create `
  --firstname "Jean" `
  --lastname "Martin" `
  --email "jean.martin@example.com" `
  --phone "+33612345678" `
  --address "24 Avenue des Fleurs" `
  --zip "69001" `
  --city "Lyon" `
  --country "FR" `
  -o json
```

### Consulter, modifier ou supprimer un contact
```powershell
./bin/abby.exe contacts get <contactId> -o json
./bin/abby.exe contacts update <contactId> --firstname "Jean" --lastname "Martin" --email "jean@example.com" -o json
./bin/abby.exe contacts delete <contactId>
```

---

## Bonnes pratiques pour la facturation

- **Prérequis de finalisation** : Une facture ne pourra être finalisée (`finalizable: true`) que si le client possède une adresse de facturation complète (`address`, `zipCode`, `city`, `country`). Vérifiez toujours l'adresse du client si `finalizeRequirements` remonte une alerte.
