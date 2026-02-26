package cmd

import (
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

var webAddr string

func init() {
	webCmd.Flags().StringVar(&webAddr, "addr", ":8080", "web server listen address")
	rootCmd.AddCommand(webCmd)
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "start web UI for feed_sources and user_subscriptions",
	Run: func(cmd *cobra.Command, args []string) {
		h := newWebHandler()
		srv := &http.Server{
			Addr:              webAddr,
			Handler:           h,
			ReadHeaderTimeout: 5 * time.Second,
		}
		cmd.Printf("web ui running at http://127.0.0.1%s\n", webAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cmd.Printf("web server failed: %v\n", err)
		}
	},
}

type dashboardData struct {
	Message           string
	Error             string
	FeedSources       []models.FeedSource
	UserSubscriptions []models.UserSubscription
	FeedForm          feedFormData
	UserForm          userFormData
}

type feedFormData struct {
	ID           uint64
	Subscription string
	Name         string
	FeedURL      string
	Language     string
	ContentField string
	ScheduleType string
	CronSpec     string
	IsEditing    bool
}

type userFormData struct {
	ID             uint64
	Email          string
	SubscriptionID string
	IsEditing      bool
}

func newWebHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleDashboard)
	mux.HandleFunc("POST /feed-sources", handleFeedSourceUpsert)
	mux.HandleFunc("POST /feed-sources/delete", handleFeedSourceDelete)
	mux.HandleFunc("POST /user-subscriptions", handleUserSubscriptionUpsert)
	mux.HandleFunc("POST /user-subscriptions/delete", handleUserSubscriptionDelete)
	return mux
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	feedSources, err := models.NewFeedSourceDao().ListAll(ctx)
	if err != nil {
		renderDashboard(w, http.StatusInternalServerError, dashboardData{Error: err.Error()})
		return
	}
	sort.Slice(feedSources, func(i, j int) bool {
		return feedSources[i].ID > feedSources[j].ID
	})
	userSubscriptions, err := models.NewUserSubscriptionDao().ListAll(ctx)
	if err != nil {
		renderDashboard(w, http.StatusInternalServerError, dashboardData{Error: err.Error()})
		return
	}
	feedForm := feedFormData{
		Language:     "auto",
		ContentField: string(constants.FeedContentFieldDescription),
		ScheduleType: string(constants.ScheduleTypeLive),
	}
	userForm := userFormData{}
	renderDashboard(w, http.StatusOK, dashboardData{
		Message:           r.URL.Query().Get("msg"),
		Error:             r.URL.Query().Get("err"),
		FeedSources:       feedSources,
		UserSubscriptions: userSubscriptions,
		FeedForm:          feedForm,
		UserForm:          userForm,
	})
}

func handleFeedSourceUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectError(w, r, "parse form failed: "+err.Error())
		return
	}
	subscriptionID := constants.SubscriptionID(strings.TrimSpace(r.FormValue("subscription_id")))
	name := strings.TrimSpace(r.FormValue("name"))
	feedURL := strings.TrimSpace(r.FormValue("feed_url"))
	language := normalizeFeedLanguage(strings.TrimSpace(r.FormValue("language")))
	contentField := constants.FeedContentField(strings.ToLower(strings.TrimSpace(r.FormValue("content_field"))))
	scheduleType := constants.ScheduleType(strings.ToLower(strings.TrimSpace(r.FormValue("schedule_type"))))
	cronSpec := strings.TrimSpace(r.FormValue("cron_spec"))
	editID, _ := strconv.ParseUint(strings.TrimSpace(r.FormValue("id")), 10, 64)

	if subscriptionID == "" || name == "" || feedURL == "" {
		redirectError(w, r, "subscription_id, name, feed_url cannot be empty")
		return
	}
	if contentField == "" {
		contentField = constants.FeedContentFieldDescription
	}
	if language == "" {
		language = "auto"
	}
	if scheduleType == "" {
		scheduleType = constants.ScheduleTypeLive
	}
	if contentField != constants.FeedContentFieldDescription && contentField != constants.FeedContentFieldContent {
		redirectError(w, r, "content_field must be description or content")
		return
	}
	if scheduleType != constants.ScheduleTypeStartup && scheduleType != constants.ScheduleTypeLive && scheduleType != constants.ScheduleTypeCron {
		redirectError(w, r, "schedule_type must be startup/live/cron")
		return
	}
	if scheduleType == constants.ScheduleTypeCron {
		if cronSpec == "" {
			redirectError(w, r, "cron_spec is required when schedule_type=cron")
			return
		}
		if _, err := cron.ParseStandard(cronSpec); err != nil {
			redirectError(w, r, "invalid cron_spec: "+err.Error())
			return
		}
	} else {
		cronSpec = ""
	}
	if editID > 0 {
		updates := map[string]interface{}{
			"subscription_id": subscriptionID,
			"name":            name,
			"feed_url":        feedURL,
			"language":        language,
			"content_field":   contentField,
			"schedule_type":   scheduleType,
			"cron_spec":       cronSpec,
			"deleted":         0,
			"updated_at":      time.Now(),
		}
		if err := models.NewFeedSourceDao().UpdateByID(r.Context(), editID, updates); err != nil {
			redirectError(w, r, "update feed source failed: "+err.Error())
			return
		}
		redirectMessage(w, r, "feed source updated")
		return
	}
	feedSource := &models.FeedSource{
		Subscription: subscriptionID,
		Name:         name,
		FeedURL:      feedURL,
		Language:     language,
		ContentField: contentField,
		ScheduleType: scheduleType,
		CronSpec:     cronSpec,
		Deleted:      0,
	}
	if err := models.NewFeedSourceDao().Upsert(r.Context(), feedSource); err != nil {
		redirectError(w, r, "save feed source failed: "+err.Error())
		return
	}
	redirectMessage(w, r, "feed source saved")
}

func handleFeedSourceDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectError(w, r, "parse form failed: "+err.Error())
		return
	}
	id, err := strconv.ParseUint(strings.TrimSpace(r.FormValue("id")), 10, 64)
	if err != nil || id == 0 {
		redirectError(w, r, "invalid feed source id")
		return
	}
	if err := models.NewFeedSourceDao().SoftDeleteByID(r.Context(), id); err != nil {
		redirectError(w, r, "delete feed source failed: "+err.Error())
		return
	}
	redirectMessage(w, r, "feed source deleted")
}

func handleUserSubscriptionUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectError(w, r, "parse form failed: "+err.Error())
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	subscriptionID := constants.SubscriptionID(strings.TrimSpace(r.FormValue("subscription_id")))
	process := strings.TrimSpace(r.FormValue("process"))
	editID, _ := strconv.ParseUint(strings.TrimSpace(r.FormValue("id")), 10, 64)
	if email == "" || subscriptionID == "" {
		redirectError(w, r, "email and subscription_id cannot be empty")
		return
	}
	if !emailRegex.MatchString(email) {
		redirectError(w, r, "invalid email")
		return
	}
	exists, err := models.NewFeedSourceDao().ExistsBySubscriptionID(r.Context(), subscriptionID)
	if err != nil {
		redirectError(w, r, "subscription check failed: "+err.Error())
		return
	}
	if !exists {
		redirectError(w, r, "subscription_id not found in feed_sources")
		return
	}

	item := &models.UserSubscription{
		Email:            email,
		SubscriptionID:   subscriptionID,
		SubscriptionType: constants.SubscriptionTypeRss,
		Process:          process,
		ProcessType:      constants.ProcessTypeGUID,
		Deleted:          0,
	}
	if editID > 0 {
		updates := map[string]interface{}{
			"email":             item.Email,
			"subscription_id":   item.SubscriptionID,
			"subscription_type": item.SubscriptionType,
			"process":           item.Process,
			"deleted":           0,
			"updated_at":        time.Now(),
		}
		if err := models.NewUserSubscriptionDao().Update(r.Context(), editID, updates); err != nil {
			redirectError(w, r, "update user subscription failed: "+err.Error())
			return
		}
		redirectMessage(w, r, "user subscription updated")
		return
	}
	if err := models.NewUserSubscriptionDao().UpsertByUnique(r.Context(), item); err != nil {
		redirectError(w, r, "save user subscription failed: "+err.Error())
		return
	}
	redirectMessage(w, r, "user subscription saved")
}

func handleUserSubscriptionDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectError(w, r, "parse form failed: "+err.Error())
		return
	}
	id, err := strconv.ParseUint(strings.TrimSpace(r.FormValue("id")), 10, 64)
	if err != nil || id == 0 {
		redirectError(w, r, "invalid user subscription id")
		return
	}
	if err := models.NewUserSubscriptionDao().SoftDeleteByID(r.Context(), id); err != nil {
		redirectError(w, r, "delete user subscription failed: "+err.Error())
		return
	}
	redirectMessage(w, r, "user subscription deleted")
}

func renderDashboard(w http.ResponseWriter, status int, data dashboardData) {
	t := template.Must(template.New("dashboard").Parse(dashboardTpl))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = t.Execute(w, data)
}

func redirectMessage(w http.ResponseWriter, r *http.Request, msg string) {
	redirectToDashboard(w, r, "msg", msg)
}

func redirectError(w http.ResponseWriter, r *http.Request, err string) {
	redirectToDashboard(w, r, "err", err)
}

func redirectToDashboard(w http.ResponseWriter, r *http.Request, key, value string) {
	q := url.Values{}
	q.Set(key, value)
	http.Redirect(w, r, "/?"+q.Encode(), http.StatusSeeOther)
}

const dashboardTpl = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>RSS2Email 配置面板</title>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Fira+Sans:wght@400;500;600;700&family=IBM+Plex+Mono:wght@400;500&display=swap');
    :root {
      --bg: #f4f7fb;
      --bg-elev: #ffffff;
      --bg-soft: #eef3f9;
      --text: #122034;
      --muted: #5f7088;
      --line: #d7e0ec;
      --brand: #0f5bd8;
      --brand-strong: #0a45a8;
      --danger: #b42318;
      --danger-bg: #fff1f0;
      --success: #067647;
      --success-bg: #ecfdf3;
      --radius: 14px;
      --shadow: 0 12px 32px rgba(17, 31, 53, 0.08);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Fira Sans", "PingFang SC", "Microsoft YaHei", sans-serif;
      color: var(--text);
      background:
        radial-gradient(900px 500px at 100% -10%, #d8e7ff 0%, rgba(216, 231, 255, 0) 60%),
        radial-gradient(760px 360px at 0% -20%, #e7f8ef 0%, rgba(231, 248, 239, 0) 58%),
        var(--bg);
      padding: 28px;
    }
    .layout { max-width: 1320px; margin: 0 auto; display: grid; gap: 20px; }
    .hero {
      background: linear-gradient(120deg, #0f172a 0%, #0b254e 52%, #113b7f 100%);
      border-radius: 18px;
      padding: 22px 24px;
      color: #f8fbff;
      box-shadow: var(--shadow);
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 16px;
      flex-wrap: wrap;
    }
    .hero h1 { margin: 0; font-size: 34px; line-height: 1.15; letter-spacing: 0.2px; }
    .hero p { margin: 8px 0 0; opacity: 0.9; color: #c4d5ef; }
    .hero .meta { font-family: "IBM Plex Mono", monospace; font-size: 13px; color: #b8c9e7; }
    .status {
      border-radius: 12px;
      padding: 12px 14px;
      border: 1px solid transparent;
      font-weight: 500;
    }
    .status.ok { background: var(--success-bg); color: var(--success); border-color: #b8ead2; }
    .status.err { background: var(--danger-bg); color: var(--danger); border-color: #f6c7c4; }
    .card {
      background: var(--bg-elev);
      border: 1px solid var(--line);
      border-radius: var(--radius);
      box-shadow: var(--shadow);
      padding: 18px;
      display: grid;
      gap: 14px;
    }
    .card h2 { margin: 0; font-size: 32px; letter-spacing: 0.2px; }
    .hint { margin: 0; color: var(--muted); font-size: 14px; }
    .grid-6 { display: grid; grid-template-columns: repeat(6, minmax(140px, 1fr)); gap: 10px; }
    .grid-2 { display: grid; grid-template-columns: repeat(2, minmax(180px, 1fr)); gap: 10px; }
    .field { display: grid; gap: 6px; }
    .field label { font-size: 12px; color: var(--muted); font-weight: 600; text-transform: uppercase; letter-spacing: 0.4px; }
    input, select {
      width: 100%;
      min-height: 44px;
      border: 1px solid #c4d0df;
      border-radius: 10px;
      background: #fff;
      color: var(--text);
      padding: 10px 12px;
      font-size: 15px;
      transition: border-color .2s ease, box-shadow .2s ease;
    }
    input:focus, select:focus {
      outline: none;
      border-color: var(--brand);
      box-shadow: 0 0 0 3px rgba(15, 91, 216, 0.17);
    }
    .actions { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; margin-top: 8px; }
    button, .link-btn {
      min-height: 44px;
      padding: 10px 15px;
      border-radius: 10px;
      border: 1px solid transparent;
      font-size: 15px;
      font-weight: 600;
      cursor: pointer;
      transition: transform .15s ease, background-color .2s ease, border-color .2s ease;
      text-decoration: none;
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
    button[disabled] { opacity: 0.7; cursor: not-allowed; transform: none; }
    button:hover, .link-btn:hover { transform: translateY(-1px); }
    button:focus-visible, .link-btn:focus-visible { outline: none; box-shadow: 0 0 0 3px rgba(15, 91, 216, 0.22); }
    .primary { background: var(--brand); color: #fff; }
    .primary:hover { background: var(--brand-strong); }
    .neutral { background: var(--bg-soft); color: #1f3553; border-color: #cdd8e7; }
    .danger { background: #fff; color: var(--danger); border-color: #efb9b4; }
    .danger:hover { background: #fff4f3; }
    .table-wrap { width: 100%; overflow-x: auto; border: 1px solid var(--line); border-radius: 12px; }
    table { width: 100%; min-width: 1200px; border-collapse: collapse; }
    th, td { padding: 12px 10px; text-align: left; border-top: 1px solid var(--line); vertical-align: top; }
    th {
      border-top: 0;
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: .45px;
      color: var(--muted);
      font-weight: 700;
      background: #f8fbff;
      position: sticky;
      top: 0;
    }
    td code { font-family: "IBM Plex Mono", monospace; font-size: 12px; color: #15438a; }
    tr:hover td { background: #f8fbff; }
    form.inline { display: inline-flex; margin: 0; }
    .row-hidden-form { display: none; }
    .row-actions {
      display: flex;
      align-items: center;
      justify-content: flex-end;
      flex-wrap: nowrap;
      gap: 10px;
      min-width: 0;
    }
    .row-actions form { margin: 0; }
    .row-btn {
      min-height: 36px;
      min-width: 72px;
      padding: 6px 14px;
      font-size: 14px;
      border-radius: 9px;
      line-height: 1;
      white-space: nowrap;
    }
    .action-col { width: 190px; min-width: 190px; }
    td.action-cell { vertical-align: middle; }
    .mono { font-family: "IBM Plex Mono", monospace; }
    .row-input {
      min-height: 36px;
      padding: 6px 8px;
      font-size: 13px;
      border-radius: 8px;
    }
    .row-input[disabled] {
      background: #f8fbff;
      border-color: #dce5f0;
      color: #2d3f58;
      opacity: 1;
      cursor: default;
    }
    .visually-hidden { position: absolute; left: -9999px; }
    @media (max-width: 1080px) {
      .grid-6 { grid-template-columns: repeat(2, minmax(150px, 1fr)); }
      .grid-2 { grid-template-columns: 1fr; }
      .hero h1 { font-size: 28px; }
    }
    @media (max-width: 640px) {
      body { padding: 12px; }
      .grid-6 { grid-template-columns: 1fr; }
      .card h2 { font-size: 28px; }
      .hero { padding: 16px; }
    }
  </style>
</head>
<body>
  <main class="layout">
    <section class="hero">
      <div>
        <h1>RSS2Email 控制台</h1>
        <p>集中管理订阅源与用户订阅，支持快速编辑与生效。</p>
      </div>
      <div class="meta">Config Panel · feed_sources / user_subscriptions</div>
    </section>

    {{if .Message}}<div class="status ok" role="status">{{.Message}}</div>{{end}}
    {{if .Error}}<div class="status err" role="alert">{{.Error}}</div>{{end}}

    <section class="card">
      <h2>Feed Sources</h2>
      <p class="hint">同 subscription_id 会更新原记录；如果之前被删除，会自动恢复。</p>
      <form id="feedForm" method="post" action="/feed-sources">
        <div class="grid-6">
          <div class="field">
            <label for="feed_subscription_id">Subscription ID</label>
            <input id="feed_subscription_id" name="subscription_id" value="{{.FeedForm.Subscription}}" required />
          </div>
          <div class="field">
            <label for="feed_name">Name</label>
            <input id="feed_name" name="name" value="{{.FeedForm.Name}}" required />
          </div>
          <div class="field">
            <label for="feed_url">Feed URL</label>
            <input id="feed_url" name="feed_url" value="{{.FeedForm.FeedURL}}" required />
          </div>
          <div class="field">
            <label for="feed_language">Language</label>
            <input id="feed_language" name="language" value="{{.FeedForm.Language}}" placeholder="auto/en/zh" />
          </div>
          <div class="field">
            <label for="feed_content_field">Content Field</label>
            <select id="feed_content_field" name="content_field">
              <option value="description" {{if eq .FeedForm.ContentField "description"}}selected{{end}}>description</option>
              <option value="content" {{if eq .FeedForm.ContentField "content"}}selected{{end}}>content</option>
            </select>
          </div>
          <div class="field">
            <label for="feed_schedule_type">Schedule Type</label>
            <select id="feed_schedule_type" name="schedule_type">
              <option value="live" {{if eq .FeedForm.ScheduleType "live"}}selected{{end}}>live</option>
              <option value="startup" {{if eq .FeedForm.ScheduleType "startup"}}selected{{end}}>startup</option>
              <option value="cron" {{if eq .FeedForm.ScheduleType "cron"}}selected{{end}}>cron</option>
            </select>
          </div>
          <div class="field">
            <label for="feed_cron_spec">Cron Spec</label>
            <input id="feed_cron_spec" class="mono" name="cron_spec" value="{{.FeedForm.CronSpec}}" placeholder="例如：0 */2 * * *" />
          </div>
        </div>
        <div class="actions">
          <button id="feedSubmitBtn" class="primary" type="submit">新增 Feed Source</button>
        </div>
      </form>

      <p class="hint">列表支持行内编辑。点击“编辑”后可直接改这一行，按钮会切换为“保存”。</p>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th><th>Subscription</th><th>Name</th><th>Feed URL</th><th>Language</th><th>Content</th><th>Schedule</th><th>Cron Spec</th><th>Created</th><th>Updated</th><th class="action-col">Action</th>
            </tr>
          </thead>
          <tbody>
            {{range .FeedSources}}
            <tr>
              <td>{{.ID}}</td>
              <td>
                <input class="row-input mono feed-row-input feed-row-{{.ID}}" name="subscription_id" form="feed-row-form-{{.ID}}" value="{{.Subscription}}" disabled required />
              </td>
              <td>
                <input class="row-input feed-row-input feed-row-{{.ID}}" name="name" form="feed-row-form-{{.ID}}" value="{{.Name}}" disabled required />
              </td>
              <td>
                <input class="row-input mono feed-row-input feed-row-{{.ID}}" name="feed_url" form="feed-row-form-{{.ID}}" value="{{.FeedURL}}" disabled required />
              </td>
              <td>
                <input class="row-input mono feed-row-input feed-row-{{.ID}}" name="language" form="feed-row-form-{{.ID}}" value="{{.Language}}" disabled />
              </td>
              <td>
                <select class="row-input feed-row-input feed-row-{{.ID}}" name="content_field" form="feed-row-form-{{.ID}}" disabled>
                  <option value="description" {{if eq .ContentField "description"}}selected{{end}}>description</option>
                  <option value="content" {{if eq .ContentField "content"}}selected{{end}}>content</option>
                </select>
              </td>
              <td>
                <select class="row-input feed-row-input feed-row-{{.ID}} feed-row-schedule" data-row-id="{{.ID}}" name="schedule_type" form="feed-row-form-{{.ID}}" disabled>
                  <option value="live" {{if eq .ScheduleType "live"}}selected{{end}}>live</option>
                  <option value="startup" {{if eq .ScheduleType "startup"}}selected{{end}}>startup</option>
                  <option value="cron" {{if eq .ScheduleType "cron"}}selected{{end}}>cron</option>
                </select>
              </td>
              <td>
                <input class="row-input mono feed-row-input feed-row-{{.ID}} feed-row-cron" data-row-id="{{.ID}}" name="cron_spec" form="feed-row-form-{{.ID}}" value="{{.CronSpec}}" disabled />
              </td>
              <td class="mono">{{.CreatedAt.Format "2006-01-02 15:04:05"}}</td>
              <td class="mono">{{.UpdatedAt.Format "2006-01-02 15:04:05"}}</td>
              <td class="action-cell">
                <form id="feed-row-form-{{.ID}}" class="row-hidden-form" method="post" action="/feed-sources">
                  <input type="hidden" name="id" value="{{.ID}}" />
                </form>
                <div class="row-actions">
                  <button class="row-btn neutral feed-row-edit-toggle" type="button" data-row-id="{{.ID}}">编辑</button>
                  <button class="row-btn neutral feed-row-cancel visually-hidden" type="button" data-row-id="{{.ID}}">取消</button>
                  <form class="inline" method="post" action="/feed-sources/delete" onsubmit="return confirm('确认删除该 feed source 吗？')">
                    <input type="hidden" name="id" value="{{.ID}}" />
                    <button class="row-btn danger" type="submit">删除</button>
                  </form>
                </div>
              </td>
            </tr>
            {{else}}
            <tr><td colspan="11">暂无数据</td></tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </section>

    <section class="card">
      <h2>User Subscriptions</h2>
      <p class="hint">subscription_id 必须已存在于 feed_sources。</p>
      <form id="userForm" method="post" action="/user-subscriptions">
        <div class="grid-2">
          <div class="field">
            <label for="user_email">Email</label>
            <input id="user_email" name="email" type="email" value="{{.UserForm.Email}}" required />
          </div>
          <div class="field">
            <label for="user_subscription_id">Subscription ID</label>
            <input id="user_subscription_id" name="subscription_id" value="{{.UserForm.SubscriptionID}}" required />
          </div>
        </div>
        <div class="actions">
          <button id="userSubmitBtn" class="primary" type="submit">新增 User Subscription</button>
        </div>
      </form>

      <p class="hint">列表支持行内编辑邮箱与订阅 ID，更多字段会完整展示。</p>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th><th>Email</th><th>Subscription</th><th>Subscription Type</th><th>Process Type</th><th>Process</th><th>Created</th><th>Updated</th><th class="action-col">Action</th>
            </tr>
          </thead>
          <tbody>
            {{range .UserSubscriptions}}
            <tr>
              <td>{{.ID}}</td>
              <td>
                <input class="row-input user-row-input user-row-{{.ID}}" name="email" form="user-row-form-{{.ID}}" value="{{.Email}}" disabled required />
              </td>
              <td>
                <input class="row-input mono user-row-input user-row-{{.ID}}" name="subscription_id" form="user-row-form-{{.ID}}" value="{{.SubscriptionID}}" disabled required />
              </td>
              <td>{{.SubscriptionType}}</td>
              <td><span class="mono">{{.ProcessType}}</span></td>
              <td>
                <input class="row-input mono user-row-input user-row-{{.ID}}" name="process" form="user-row-form-{{.ID}}" value="{{.Process}}" disabled />
              </td>
              <td class="mono">{{.CreatedAt.Format "2006-01-02 15:04:05"}}</td>
              <td class="mono">{{.UpdatedAt.Format "2006-01-02 15:04:05"}}</td>
              <td class="action-cell">
                <form id="user-row-form-{{.ID}}" class="row-hidden-form" method="post" action="/user-subscriptions">
                  <input type="hidden" name="id" value="{{.ID}}" />
                </form>
                <div class="row-actions">
                  <button class="row-btn neutral user-row-edit-toggle" type="button" data-row-id="{{.ID}}">编辑</button>
                  <button class="row-btn neutral user-row-cancel visually-hidden" type="button" data-row-id="{{.ID}}">取消</button>
                  <form class="inline" method="post" action="/user-subscriptions/delete" onsubmit="return confirm('确认删除该用户订阅吗？')">
                    <input type="hidden" name="id" value="{{.ID}}" />
                    <button class="row-btn danger" type="submit">删除</button>
                  </form>
                </div>
              </td>
            </tr>
            {{else}}
            <tr><td colspan="9">暂无数据</td></tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </section>
  </main>
  <script>
    (function () {
      const feedForm = document.getElementById('feedForm');
      const userForm = document.getElementById('userForm');
      if (!feedForm || !userForm) return;

      const feedSubscription = document.getElementById('feed_subscription_id');
      const feedScheduleType = document.getElementById('feed_schedule_type');
      const feedCronSpec = document.getElementById('feed_cron_spec');
      const feedSubmitBtn = document.getElementById('feedSubmitBtn');
      const userSubmitBtn = document.getElementById('userSubmitBtn');

      function updateFeedCronState() {
        const isCron = feedScheduleType.value === 'cron';
        feedCronSpec.required = isCron;
        feedCronSpec.disabled = !isCron;
        if (!isCron) {
          feedCronSpec.value = '';
          feedCronSpec.placeholder = 'schedule_type=cron 时填写';
        } else {
          feedCronSpec.placeholder = '例如：0 */2 * * *';
        }
      }

      function updateFeedRowCronState(rowID) {
        const schedule = document.querySelector('.feed-row-schedule[data-row-id="' + rowID + '"]');
        const cron = document.querySelector('.feed-row-cron[data-row-id="' + rowID + '"]');
        if (!schedule || !cron) return;
        const isCron = schedule.value === 'cron';
        cron.required = isCron;
        cron.disabled = !isCron;
        if (!isCron) cron.value = '';
      }

      function setRowEditable(prefix, rowID, editable) {
        document.querySelectorAll('.' + prefix + '-row-' + rowID).forEach((el) => {
          el.disabled = !editable;
        });
      }

      document.querySelectorAll('.feed-row-edit-toggle').forEach((btn) => {
        btn.addEventListener('click', () => {
          const rowID = btn.dataset.rowId;
          const cancelBtn = document.querySelector('.feed-row-cancel[data-row-id="' + rowID + '"]');
          const editing = btn.dataset.editing === '1';
          if (!editing) {
            setRowEditable('feed', rowID, true);
            updateFeedRowCronState(rowID);
            btn.dataset.editing = '1';
            btn.textContent = '保存';
            btn.classList.remove('neutral');
            btn.classList.add('primary');
            if (cancelBtn) cancelBtn.classList.remove('visually-hidden');
            const first = document.querySelector('.feed-row-' + rowID);
            if (first) first.focus();
            return;
          }
          const form = document.getElementById('feed-row-form-' + rowID);
          if (form) form.requestSubmit();
        });
      });

      document.querySelectorAll('.feed-row-cancel').forEach((btn) => {
        btn.addEventListener('click', () => {
          const rowID = btn.dataset.rowId;
          window.location.reload();
        });
      });

      document.querySelectorAll('.feed-row-schedule').forEach((el) => {
        el.addEventListener('change', () => updateFeedRowCronState(el.dataset.rowId));
      });

      feedScheduleType.addEventListener('change', updateFeedCronState);

      document.querySelectorAll('.user-row-edit-toggle').forEach((btn) => {
        btn.addEventListener('click', () => {
          const rowID = btn.dataset.rowId;
          const cancelBtn = document.querySelector('.user-row-cancel[data-row-id="' + rowID + '"]');
          const editing = btn.dataset.editing === '1';
          if (!editing) {
            setRowEditable('user', rowID, true);
            btn.dataset.editing = '1';
            btn.textContent = '保存';
            btn.classList.remove('neutral');
            btn.classList.add('primary');
            if (cancelBtn) cancelBtn.classList.remove('visually-hidden');
            const first = document.querySelector('.user-row-' + rowID);
            if (first) first.focus();
            return;
          }
          const form = document.getElementById('user-row-form-' + rowID);
          if (form) form.requestSubmit();
        });
      });

      document.querySelectorAll('.user-row-cancel').forEach((btn) => {
        btn.addEventListener('click', () => {
          window.location.reload();
        });
      });

      // Enforce locked state for inline rows until the user clicks "编辑".
      document.querySelectorAll('.feed-row-input, .user-row-input').forEach((el) => {
        el.disabled = true;
      });
      document.querySelectorAll('.feed-row-edit-toggle, .user-row-edit-toggle').forEach((btn) => {
        btn.dataset.editing = '0';
        btn.textContent = '编辑';
        btn.classList.remove('primary');
        btn.classList.add('neutral');
      });
      document.querySelectorAll('.feed-row-cancel, .user-row-cancel').forEach((btn) => {
        btn.classList.add('visually-hidden');
      });

      [feedForm, userForm].forEach((form) => {
        form.addEventListener('submit', (e) => {
          const submit = form.querySelector('button[type="submit"]');
          if (!submit) return;
          submit.disabled = true;
          submit.dataset.originalText = submit.textContent;
          submit.textContent = '保存中...';
          window.setTimeout(() => {
            submit.disabled = false;
            submit.textContent = submit.dataset.originalText || submit.textContent;
          }, 3000);
        });
      });

      feedSubmitBtn.textContent = '新增 Feed Source';
      userSubmitBtn.textContent = '新增 User Subscription';
      updateFeedCronState();
    })();
  </script>
</body>
</html>`
