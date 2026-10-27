package app

import (
	"database/sql"
	"maps"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/database"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramControlRequestCapacity returns durable diagnostics for records,
// authoring, work admission and confirmations at the ordinary request ceiling.
// None can create an artifact or consume its existing confirmation.
func TestTelegramControlRequestCapacity(t *testing.T) {
	f := telegramOperationAppFixture(t)
	action, _ := telegramOperationOfferRecord(t, f)
	f.send(t, "/prepare --project example-project Capacity preparation", 0)

	c := f.sessions(t)[0]
	f.send(t, "/reply "+c.ID+" Concrete requirements.", 0)

	operations := telegramOperationConcreteService(t, f.receiver)

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("capacity project unavailable", projects, err)
	}

	seedTelegramAppRequestCapacity(t, f, action.RequestID)

	images := workflowCLIImage(t, projects[0].Path)
	for _, command := range []string{"/answer " + action.Token + " confirm", `/tickets create {"title":"Capacity rejected","synopsis":"No effect"}`, "/done " + c.ID, "/work ticket PRJ-1"} {
		if text := f.send(t, command, 0).Text; !strings.Contains(text, "The request queue is full. This request was not accepted. Retry after pending requests finish; existing work is unchanged.") {
			t.Fatal("capacity failed without its durable diagnostic", text)
		}
	}

	if !maps.Equal(images, workflowCLIImage(t, projects[0].Path)) {
		t.Fatal("request capacity rejection changed a project artifact")
	}

	checkTelegramCapacityAction(t, f, action)

	if text := f.send(t, "/status", 0).Text; !strings.Contains(text, "Receiver:") {
		t.Fatal("request capacity blocked status", text)
	}
}

func checkTelegramCapacityAction(t *testing.T, f *telegramConversationAppFixture, action telegramcore.TelegramAction) {
	t.Helper()

	request, err := f.receiver.store.TelegramActionRequest(t.Context(), action.Peer, telegramcore.TelegramActionAnswer{Token: action.Token, Choice: "confirm"})
	if err != nil {
		t.Fatal(err)
	}

	retained, err := f.receiver.store.TelegramRequestAction(t.Context(), request)
	if err != nil || retained.State != "pending" || retained.AnswerRequestID != "" {
		t.Fatal("capacity consumed an unaccepted action", retained, err)
	}
}

func seedTelegramAppRequestCapacity(t *testing.T, f *telegramConversationAppFixture, template string) {
	t.Helper()

	db, err := sql.Open("sqlite", database.UserDatabasePath(f.receiver.env))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	var present int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_requests`).Scan(&present); err != nil {
		t.Fatal(err)
	}

	count := 960 - present

	_, err = db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<?)
INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at)
SELECT u.bot_id,100000+n.x,u.cursor_epoch,u.payload_digest,'message',u.chat_id,u.topic_id,u.user_id,n.x,u.chat_type,u.received_at,'','processed','capacity_fixture',u.updated_at FROM n JOIN telegram_requests r ON r.id=? JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id`, count, template)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<?)
INSERT INTO telegram_requests(id,bot_id,update_id,operation_ordinal,operator_id,project_id,project_binding_digest,namespace,operation_id,arguments,argument_digest,state,created_at,updated_at)
SELECT 'capacity-'||n.x,r.bot_id,100000+n.x,0,r.operator_id,r.project_id,r.project_binding_digest,r.namespace,r.operation_id,r.arguments,r.argument_digest,'prepared',r.created_at,r.updated_at FROM n JOIN telegram_requests r ON r.id=?`, count, template)
	if err != nil {
		t.Fatal(err)
	}
}
