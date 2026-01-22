-- +goose Up
-- +goose StatementBegin
-- Add locale column to email_template table
ALTER TABLE email_template 
  ADD COLUMN IF NOT EXISTS locale VARCHAR(10) DEFAULT 'en';

-- Update unique constraint to include locale
ALTER TABLE email_template 
  DROP CONSTRAINT IF EXISTS unique_template_name_per_account;

ALTER TABLE email_template 
  ADD CONSTRAINT unique_template_name_per_account_locale 
  UNIQUE (account_id, name, locale);

-- Create index for locale
CREATE INDEX IF NOT EXISTS idx_email_template_locale ON email_template(locale);

-- Migrate existing templates to have locale 'en' explicitly
UPDATE email_template SET locale = 'en' WHERE locale IS NULL OR locale = '';

-- Insert translations for all supported languages
-- Base template translations (same for all locales, but we'll create entries for consistency)
INSERT INTO email_template (name, locale, content, is_system) 
SELECT 'base', 'ru', content, TRUE
FROM email_template 
WHERE name = 'base' AND locale = 'en'
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

INSERT INTO email_template (name, locale, content, is_system) 
SELECT 'base', 'es', content, TRUE
FROM email_template 
WHERE name = 'base' AND locale = 'en'
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

INSERT INTO email_template (name, locale, content, is_system) 
SELECT 'base', 'fr', content, TRUE
FROM email_template 
WHERE name = 'base' AND locale = 'en'
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

INSERT INTO email_template (name, locale, content, is_system) 
SELECT 'base', 'de', content, TRUE
FROM email_template 
WHERE name = 'base' AND locale = 'en'
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

INSERT INTO email_template (name, locale, content, is_system) 
SELECT 'base', 'it', content, TRUE
FROM email_template 
WHERE name = 'base' AND locale = 'en'
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

INSERT INTO email_template (name, locale, content, is_system) 
SELECT 'base', 'pt', content, TRUE
FROM email_template 
WHERE name = 'base' AND locale = 'en'
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

-- Invitation template translations
INSERT INTO email_template (name, locale, subject, content, is_system) VALUES
('invitation', 'ru', 'Вам предложено подписать документ', '{{define "content"}}
<p>Здравствуйте, {{.RecipientName}},</p>

<p>Вам предложено подписать документ: <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Пожалуйста, нажмите кнопку ниже, чтобы просмотреть и подписать документ:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Подписать документ</a>
</p>

{{if .ExpiresAt}}
<p><small>Это приглашение истекает {{.ExpiresAt}}</small></p>
{{end}}

<p>Если у вас есть вопросы, пожалуйста, свяжитесь с {{.SenderName}}.</p>
{{end}}', TRUE),
('invitation', 'es', 'Has sido invitado a firmar un documento', '{{define "content"}}
<p>Hola {{.RecipientName}},</p>

<p>Has sido invitado a firmar un documento: <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Por favor, haz clic en el botón de abajo para revisar y firmar el documento:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Firmar documento</a>
</p>

{{if .ExpiresAt}}
<p><small>Esta invitación expira el {{.ExpiresAt}}</small></p>
{{end}}

<p>Si tienes alguna pregunta, por favor contacta a {{.SenderName}}.</p>
{{end}}', TRUE),
('invitation', 'fr', 'Vous avez été invité à signer un document', '{{define "content"}}
<p>Bonjour {{.RecipientName}},</p>

<p>Vous avez été invité à signer un document : <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Veuillez cliquer sur le bouton ci-dessous pour examiner et signer le document :</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Signer le document</a>
</p>

{{if .ExpiresAt}}
<p><small>Cette invitation expire le {{.ExpiresAt}}</small></p>
{{end}}

<p>Si vous avez des questions, veuillez contacter {{.SenderName}}.</p>
{{end}}', TRUE),
('invitation', 'de', 'Sie wurden eingeladen, ein Dokument zu unterschreiben', '{{define "content"}}
<p>Hallo {{.RecipientName}},</p>

<p>Sie wurden eingeladen, ein Dokument zu unterschreiben: <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Bitte klicken Sie auf die Schaltfläche unten, um das Dokument zu überprüfen und zu unterschreiben:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Dokument unterschreiben</a>
</p>

{{if .ExpiresAt}}
<p><small>Diese Einladung läuft am {{.ExpiresAt}} ab</small></p>
{{end}}

<p>Wenn Sie Fragen haben, wenden Sie sich bitte an {{.SenderName}}.</p>
{{end}}', TRUE),
('invitation', 'it', 'Sei stato invitato a firmare un documento', '{{define "content"}}
<p>Ciao {{.RecipientName}},</p>

<p>Sei stato invitato a firmare un documento: <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Fai clic sul pulsante qui sotto per rivedere e firmare il documento:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Firma documento</a>
</p>

{{if .ExpiresAt}}
<p><small>Questo invito scade il {{.ExpiresAt}}</small></p>
{{end}}

<p>Se hai domande, contatta {{.SenderName}}.</p>
{{end}}', TRUE),
('invitation', 'pt', 'Você foi convidado a assinar um documento', '{{define "content"}}
<p>Olá {{.RecipientName}},</p>

<p>Você foi convidado a assinar um documento: <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Por favor, clique no botão abaixo para revisar e assinar o documento:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Assinar documento</a>
</p>

{{if .ExpiresAt}}
<p><small>Este convite expira em {{.ExpiresAt}}</small></p>
{{end}}

<p>Se você tiver alguma dúvida, entre em contato com {{.SenderName}}.</p>
{{end}}', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

-- Reminder template translations
INSERT INTO email_template (name, locale, subject, content, is_system) VALUES
('reminder', 'ru', 'Напоминание: Документ ожидает вашей подписи', '{{define "content"}}
<p>Здравствуйте, {{.RecipientName}},</p>

<p>Это напоминание о том, что у вас есть документ, ожидающий подписи: <strong>{{.DocumentName}}</strong></p>

<p>Пожалуйста, нажмите кнопку ниже, чтобы завершить подписание:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Подписать документ</a>
</p>

{{if .ExpiresAt}}
<p><small>Это приглашение истекает {{.ExpiresAt}}</small></p>
{{end}}

<p>Спасибо за внимание.</p>
{{end}}', TRUE),
('reminder', 'es', 'Recordatorio: Documento pendiente de su firma', '{{define "content"}}
<p>Hola {{.RecipientName}},</p>

<p>Este es un recordatorio de que tienes un documento pendiente de firmar: <strong>{{.DocumentName}}</strong></p>

<p>Por favor, haz clic en el botón de abajo para completar tu firma:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Firmar documento</a>
</p>

{{if .ExpiresAt}}
<p><small>Esta invitación expira el {{.ExpiresAt}}</small></p>
{{end}}

<p>Gracias por su atención.</p>
{{end}}', TRUE),
('reminder', 'fr', 'Rappel : Document en attente de votre signature', '{{define "content"}}
<p>Bonjour {{.RecipientName}},</p>

<p>Ceci est un rappel que vous avez un document en attente de signature : <strong>{{.DocumentName}}</strong></p>

<p>Veuillez cliquer sur le bouton ci-dessous pour compléter votre signature :</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Signer le document</a>
</p>

{{if .ExpiresAt}}
<p><small>Cette invitation expire le {{.ExpiresAt}}</small></p>
{{end}}

<p>Merci pour votre attention.</p>
{{end}}', TRUE),
('reminder', 'de', 'Erinnerung: Dokument wartet auf Ihre Unterschrift', '{{define "content"}}
<p>Hallo {{.RecipientName}},</p>

<p>Dies ist eine Erinnerung, dass Sie ein ausstehendes Dokument zum Unterschreiben haben: <strong>{{.DocumentName}}</strong></p>

<p>Bitte klicken Sie auf die Schaltfläche unten, um Ihre Unterschrift zu vervollständigen:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Dokument unterschreiben</a>
</p>

{{if .ExpiresAt}}
<p><small>Diese Einladung läuft am {{.ExpiresAt}} ab</small></p>
{{end}}

<p>Vielen Dank für Ihre Aufmerksamkeit.</p>
{{end}}', TRUE),
('reminder', 'it', 'Promemoria: Documento in attesa della tua firma', '{{define "content"}}
<p>Ciao {{.RecipientName}},</p>

<p>Questo è un promemoria che hai un documento in attesa di firma: <strong>{{.DocumentName}}</strong></p>

<p>Fai clic sul pulsante qui sotto per completare la tua firma:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Firma documento</a>
</p>

{{if .ExpiresAt}}
<p><small>Questo invito scade il {{.ExpiresAt}}</small></p>
{{end}}

<p>Grazie per l''attenzione.</p>
{{end}}', TRUE),
('reminder', 'pt', 'Lembrete: Documento aguardando sua assinatura', '{{define "content"}}
<p>Olá {{.RecipientName}},</p>

<p>Este é um lembrete de que você tem um documento pendente para assinar: <strong>{{.DocumentName}}</strong></p>

<p>Por favor, clique no botão abaixo para completar sua assinatura:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Assinar documento</a>
</p>

{{if .ExpiresAt}}
<p><small>Este convite expira em {{.ExpiresAt}}</small></p>
{{end}}

<p>Obrigado pela atenção.</p>
{{end}}', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;

-- Completed template translations
INSERT INTO email_template (name, locale, subject, content, is_system) VALUES
('completed', 'ru', 'Документ успешно подписан', '{{define "content"}}
<p>Здравствуйте, {{.RecipientName}},</p>

<p>Документ <strong>{{.DocumentName}}</strong> был завершен и подписан всеми сторонами.</p>

<p>Вы можете скачать завершенный документ из вашей панели управления.</p>

<p>Спасибо за использование нашего сервиса.</p>
{{end}}', TRUE),
('completed', 'es', 'Documento firmado exitosamente', '{{define "content"}}
<p>Hola {{.RecipientName}},</p>

<p>El documento <strong>{{.DocumentName}}</strong> ha sido completado y firmado por todas las partes.</p>

<p>Puedes descargar el documento completado desde tu panel de control.</p>

<p>Gracias por usar nuestro servicio.</p>
{{end}}', TRUE),
('completed', 'fr', 'Document signé avec succès', '{{define "content"}}
<p>Bonjour {{.RecipientName}},</p>

<p>Le document <strong>{{.DocumentName}}</strong> a été complété et signé par toutes les parties.</p>

<p>Vous pouvez télécharger le document complété depuis votre tableau de bord.</p>

<p>Merci d''utiliser notre service.</p>
{{end}}', TRUE),
('completed', 'de', 'Dokument erfolgreich unterschrieben', '{{define "content"}}
<p>Hallo {{.RecipientName}},</p>

<p>Das Dokument <strong>{{.DocumentName}}</strong> wurde von allen Parteien abgeschlossen und unterschrieben.</p>

<p>Sie können das abgeschlossene Dokument von Ihrem Dashboard herunterladen.</p>

<p>Vielen Dank für die Nutzung unseres Dienstes.</p>
{{end}}', TRUE),
('completed', 'it', 'Documento firmato con successo', '{{define "content"}}
<p>Ciao {{.RecipientName}},</p>

<p>Il documento <strong>{{.DocumentName}}</strong> è stato completato e firmato da tutte le parti.</p>

<p>Puoi scaricare il documento completato dalla tua dashboard.</p>

<p>Grazie per aver utilizzato il nostro servizio.</p>
{{end}}', TRUE),
('completed', 'pt', 'Documento assinado com sucesso', '{{define "content"}}
<p>Olá {{.RecipientName}},</p>

<p>O documento <strong>{{.DocumentName}}</strong> foi concluído e assinado por todas as partes.</p>

<p>Você pode baixar o documento concluído do seu painel de controle.</p>

<p>Obrigado por usar nosso serviço.</p>
{{end}}', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account_locale DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_email_template_locale;
ALTER TABLE email_template DROP CONSTRAINT IF EXISTS unique_template_name_per_account_locale;
ALTER TABLE email_template ADD CONSTRAINT unique_template_name_per_account UNIQUE (account_id, name);
ALTER TABLE email_template DROP COLUMN IF EXISTS locale;
-- +goose StatementEnd
