---
name: abby-auth
description: >-
  Authentification et configuration du CLI Abby (clé API Bearer suk-***,
  vérification du compte via /v2/company/me, diagnostic des erreurs 401/403).
  Utiliser ce skill lors de la configuration initiale ou en cas d'erreur d'authentification.
---

# Skill : Authentification & Configuration Abby (`abby-auth`)

Ce skill guide la configuration et le diagnostic de l'authentification par clé API sur la plateforme Abby (`https://docs.abby.fr/api/authentification`).

## Principes de l'authentification Abby

- **Méthode** : Jeton Bearer transmis dans l'en-tête HTTP `Authorization: Bearer suk-*****`.
- **Protocole** : **HTTPS obligatoire** (`https://api.app-abby.com` ou `https://api.abby.fr`). Les requêtes HTTP sont rejetées.
- **Obtention de la clé** : Depuis le tableau de bord Abby dans *Paramètres > Intégrations* (`https://my.app-abby.com/settings/integrations`).
- **Sécurité** : Ne jamais commiter la clé `suk-***` dans Git.

## Commandes CLI disponibles

### 1. Vérifier l'état de la configuration locale
```powershell
./bin/abby.exe auth status -o json
```
 L'ordre de priorité de résolution de la clé API est :
1. Flag `--api-key <suk-...>`
2. Variable d'environnement `ABBY_API_KEY`
3. Fichier de configuration local (`~/.config/abby-cli/config.json` ou `%AppData%\abby-cli\config.json`)

### 2. Enregistrer une clé API
```powershell
./bin/abby.exe auth set-key "suk-votre_cle_secrete"
```

### 3. Tester la connexion et identifier l'entreprise (`GET /v2/company/me`)
```powershell
./bin/abby.exe auth whoami -o json
```
Retourne les informations de l'utilisateur connecté, la raison sociale de l'entreprise, son SIRET, son numéro de TVA et ses préférences.

### 4. Supprimer la clé enregistrée
```powershell
./bin/abby.exe auth clear
```

## Diagnostic des erreurs courantes

| Code HTTP | Type / Code Abby | Cause & Action |
| :--- | :--- | :--- |
| `401` | `authentication_error` / `token_failure` | Clé API absente, invalide ou révoquée. Demandez à l'utilisateur une nouvelle clé depuis *Paramètres > Intégrations*. |
| `403` | `rsc_error` / `forbidden` | La clé API ou l'abonnement Abby ne dispose pas des droits requis pour cette ressource. |
| `429` | Rate Limit | Trop de requêtes simultanées. Espacez les appels CLI. |
