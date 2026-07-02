package preview

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	imagepkg "github.com/prettyletto/omarchy-themegen/internal/image"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

type SelectionState struct {
	Mode      string
	Selected  int
	Groups    map[string]int
	Overrides map[string]int
}

func NewSelectionState() *SelectionState {
	return &SelectionState{
		Groups:    make(map[string]int),
		Overrides: make(map[string]int),
	}
}

type BrowserServer struct {
	mu            sync.Mutex
	token         string
	port          int
	listener      net.Listener
	server        *http.Server
	state         *SelectionState
	directions    []theme.Direction
	sourcePath    string
	previewDir    string
	previewPaths  map[int]string
	composedPaths map[string]string
	previewErr    error
	onUpdate      func(*SelectionState)
	done          chan struct{}
	idleTimer     *time.Timer
	idleTimeout   time.Duration
}

func NewBrowserServer(sourcePath string, directions []theme.Direction, onUpdate func(*SelectionState)) *BrowserServer {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		// Fallback: use timestamp-based token
		tokenBytes = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	token := hex.EncodeToString(tokenBytes)

	return &BrowserServer{
		token:         token,
		state:         NewSelectionState(),
		directions:    directions,
		sourcePath:    sourcePath,
		previewPaths:  make(map[int]string),
		composedPaths: make(map[string]string),
		onUpdate:      onUpdate,
		done:          make(chan struct{}),
		idleTimeout:   5 * time.Minute,
	}
}

func (s *BrowserServer) Start() (string, error) {
	s.generateDirectionPreviews()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("cannot bind browser preview: %w", err)
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/select-direction", s.handleSelectDirection)
	mux.HandleFunc("/api/select-group", s.handleSelectGroup)
	mux.HandleFunc("/api/set-mode", s.handleSetMode)
	mux.HandleFunc("/api/set-override", s.handleSetOverride)
	mux.HandleFunc("/api/clear-override", s.handleClearOverride)
	mux.HandleFunc("/api/state", s.handleGetState)
	mux.HandleFunc("/preview/direction/", s.handleDirectionPreview)
	mux.HandleFunc("/preview/current", s.handleCurrentPreview)

	s.server = &http.Server{Handler: mux}

	s.resetIdleTimer()

	go func() {
		s.server.Serve(listener)
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", s.port, s.token)
	return url, nil
}

func (s *BrowserServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}
	select {
	case <-s.done:
		// Already closed
	default:
		close(s.done)
	}
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}
	if s.previewDir != "" {
		os.RemoveAll(s.previewDir)
	}
}

func (s *BrowserServer) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/?token=%s", s.port, s.token)
}

func (s *BrowserServer) Token() string {
	return s.token
}

func (s *BrowserServer) Port() int {
	return s.port
}

func (s *BrowserServer) resetIdleTimer() {
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}
	s.idleTimer = time.AfterFunc(s.idleTimeout, func() {
		s.Stop()
	})
}

func (s *BrowserServer) checkToken(r *http.Request) bool {
	return r.URL.Query().Get("token") == s.token
}

func (s *BrowserServer) generateDirectionPreviews() {
	if len(s.directions) == 0 || s.sourcePath == "" {
		return
	}
	dir, err := os.MkdirTemp("", "omarchy-themegen-browser-preview-")
	if err != nil {
		s.previewErr = err
		return
	}
	paths, err := GenerateDirectionPreviews(dir, s.sourcePath, s.directions)
	if err != nil {
		os.RemoveAll(dir)
		s.previewErr = err
		return
	}
	s.previewDir = dir
	for i, path := range paths {
		if i < len(s.directions) {
			s.previewPaths[s.directions[i].ID] = path
		}
	}
}

func (s *BrowserServer) handleDirectionPreview(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	idText := filepath.Base(r.URL.Path)
	dirID, err := strconv.Atoi(idText)
	if err != nil {
		http.Error(w, "Invalid direction", http.StatusBadRequest)
		return
	}
	path := s.previewPaths[dirID]
	if path == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, path)
}

func (s *BrowserServer) handleCurrentPreview(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	path, err := s.currentPreviewPath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, path)
}

func (s *BrowserServer) currentPreviewPath() (string, error) {
	state := s.copyState()
	key := s.previewStateKey(state)
	if path := s.composedPaths[key]; path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	if s.previewDir == "" {
		dir, err := os.MkdirTemp("", "omarchy-themegen-browser-preview-")
		if err != nil {
			return "", err
		}
		s.previewDir = dir
	}
	tm, err := s.themeModelForState(state)
	if err != nil {
		return "", err
	}
	path := filepath.Join(s.previewDir, key+".png")
	if err := GenerateComposedPreview(path, s.sourcePath, tm); err != nil {
		return "", err
	}
	s.composedPaths[key] = path
	return path, nil
}

func (s *BrowserServer) themeModelForState(state *SelectionState) (*theme.ThemeModel, error) {
	if len(s.directions) == 0 {
		return nil, fmt.Errorf("no directions available")
	}
	imgResult := imagepkg.Validate(s.sourcePath)
	if !imgResult.Valid {
		return nil, fmt.Errorf("source image validation failed")
	}
	if state.Mode == "component-mix" {
		comp := theme.NewComposition("component-mix")
		comp.Directions = s.directions
		for _, group := range theme.AllGroups {
			dirID := state.Groups[group.ID]
			if dirID == 0 {
				dirID = selectedDirectionID(state, len(s.directions))
			}
			if err := comp.SetGroupSource(group.ID, dirID); err != nil {
				return nil, err
			}
		}
		for surface, dirID := range state.Overrides {
			if dirID > 0 {
				if err := comp.SetOverride(surface, dirID); err != nil {
					return nil, err
				}
			}
		}
		return comp.Resolve("Preview", s.sourcePath, imgResult)
	}
	dirID := selectedDirectionID(state, len(s.directions))
	return theme.NewThemeModelFromDirection("Preview", s.sourcePath, imgResult, s.directions[dirID-1])
}

func selectedDirectionID(state *SelectionState, directionCount int) int {
	if state.Selected >= 1 && state.Selected <= directionCount {
		return state.Selected
	}
	return 1
}

func (s *BrowserServer) previewStateKey(state *SelectionState) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%d", state.Mode, selectedDirectionID(state, len(s.directions)))
	groupKeys := make([]string, 0, len(state.Groups))
	for key := range state.Groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)
	for _, key := range groupKeys {
		fmt.Fprintf(h, "|g:%s=%d", key, state.Groups[key])
	}
	overrideKeys := make([]string, 0, len(state.Overrides))
	for key := range state.Overrides {
		overrideKeys = append(overrideKeys, key)
	}
	sort.Strings(overrideKeys)
	for _, key := range overrideKeys {
		fmt.Fprintf(h, "|o:%s=%d", key, state.Overrides[key])
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (s *BrowserServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	directionsJSON, err := json.Marshal(s.directions)
	if err != nil {
		http.Error(w, "Cannot encode directions", http.StatusInternalServerError)
		return
	}
	previewStatus := "generated"
	if s.previewErr != nil {
		previewStatus = s.previewErr.Error()
	}
	fmt.Fprintf(w, browserHTML, previewStatus, s.token, string(directionsJSON))
}

func (s *BrowserServer) handleSelectDirection(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) || r.Method != "POST" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	var dirID int
	if _, err := fmt.Sscanf(r.URL.Query().Get("id"), "%d", &dirID); err != nil || dirID < 1 || dirID > len(s.directions) {
		http.Error(w, fmt.Sprintf("Invalid direction: %s", r.URL.Query().Get("id")), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.state.Mode = "whole-theme"
	s.state.Selected = dirID
	s.mu.Unlock()
	if s.onUpdate != nil {
		s.onUpdate(s.copyState())
	}
	w.WriteHeader(http.StatusOK)
}

func (s *BrowserServer) handleSelectGroup(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) || r.Method != "POST" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	groupID := r.URL.Query().Get("group")
	if _, ok := groupByID(groupID); !ok {
		http.Error(w, fmt.Sprintf("Unknown group: %s", groupID), http.StatusBadRequest)
		return
	}
	var dirID int
	if _, err := fmt.Sscanf(r.URL.Query().Get("id"), "%d", &dirID); err != nil || dirID < 1 || dirID > len(s.directions) {
		http.Error(w, fmt.Sprintf("Invalid direction for group: %s", r.URL.Query().Get("id")), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.state.Mode = "component-mix"
	if s.state.Groups == nil {
		s.state.Groups = make(map[string]int)
	}
	s.state.Groups[groupID] = dirID
	s.mu.Unlock()
	if s.onUpdate != nil {
		s.onUpdate(s.copyState())
	}
	w.WriteHeader(http.StatusOK)
}

func (s *BrowserServer) handleSetMode(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) || r.Method != "POST" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	mode := r.URL.Query().Get("mode")
	if mode != "whole-theme" && mode != "component-mix" {
		http.Error(w, fmt.Sprintf("Invalid mode: %s", mode), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.state.Mode = mode
	s.mu.Unlock()
	if s.onUpdate != nil {
		s.onUpdate(s.copyState())
	}
	w.WriteHeader(http.StatusOK)
}

func (s *BrowserServer) handleSetOverride(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) || r.Method != "POST" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	surface := r.URL.Query().Get("surface")
	if !validSurface(surface) {
		http.Error(w, fmt.Sprintf("Unknown surface: %s", surface), http.StatusBadRequest)
		return
	}
	var dirID int
	if _, err := fmt.Sscanf(r.URL.Query().Get("id"), "%d", &dirID); err != nil || dirID < 1 || dirID > len(s.directions) {
		http.Error(w, fmt.Sprintf("Invalid direction: %s", r.URL.Query().Get("id")), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	if s.state.Overrides == nil {
		s.state.Overrides = make(map[string]int)
	}
	s.state.Overrides[surface] = dirID
	s.mu.Unlock()
	if s.onUpdate != nil {
		s.onUpdate(s.copyState())
	}
	w.WriteHeader(http.StatusOK)
}

func (s *BrowserServer) handleClearOverride(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) || r.Method != "POST" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	surface := r.URL.Query().Get("surface")
	s.mu.Lock()
	delete(s.state.Overrides, surface)
	s.mu.Unlock()
	if s.onUpdate != nil {
		s.onUpdate(s.copyState())
	}
	w.WriteHeader(http.StatusOK)
}

func (s *BrowserServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s.resetIdleTimer()
	state := s.copyState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mode":      state.Mode,
		"selected":  state.Selected,
		"groups":    state.Groups,
		"overrides": state.Overrides,
	})
}

func (s *BrowserServer) copyState() *SelectionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	groups := make(map[string]int)
	for k, v := range s.state.Groups {
		groups[k] = v
	}
	overrides := make(map[string]int)
	for k, v := range s.state.Overrides {
		overrides[k] = v
	}
	return &SelectionState{
		Mode:      s.state.Mode,
		Selected:  s.state.Selected,
		Groups:    groups,
		Overrides: overrides,
	}
}

func groupByID(id string) (theme.SurfaceGroup, bool) {
	return theme.GroupByID(id)
}

func validSurface(name string) bool {
	return theme.ValidSurface(name)
}

const browserHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>omarchy-themegen Preview</title>
<style>
*{box-sizing:border-box}
body{font-family:Inter,system-ui,sans-serif;background:#10151f;color:#c0caf5;margin:20px}
h1{color:#7aa2f7;margin-bottom:5px}
.info{color:#6b7280;font-size:12px;margin-bottom:15px}
.dir{display:inline-block;vertical-align:top;margin:8px;padding:10px;border:2px solid #303848;border-radius:10px;cursor:pointer;width:420px;background:#151b26}
.dir:hover,.dir.selected{border-color:#82aaff}
.dir .label{font-weight:bold;margin-bottom:8px}
.preview-img{display:block;width:100%%;aspect-ratio:16/9;object-fit:cover;border-radius:6px;border:1px solid #303848;background:#0b1020}
.current-preview{max-width:1120px;margin:12px 0 18px;padding:12px;border:1px solid #303848;border-radius:12px;background:#151b26}
.current-preview img{display:block;width:100%%;aspect-ratio:16/9;object-fit:cover;border-radius:8px;border:1px solid #303848;background:#0b1020}
.mock{position:relative;height:225px;overflow:hidden;border-radius:6px;background-size:cover;background-position:center;border:1px solid #303848;display:none}
.mock:before{content:"";position:absolute;inset:0;background:var(--bg);opacity:.74}
.bar{position:absolute;left:0;right:0;top:0;height:10px;background:var(--c8);opacity:.95}
.chrome{position:absolute;left:8px;top:18px;width:168px;height:92px;background:var(--c8);opacity:.92;border-radius:3px}
.term{position:absolute;left:8px;top:116px;width:168px;height:78px;background:var(--bg);border:1px solid var(--accent);opacity:.94;border-radius:3px}
.logo{position:absolute;left:184px;top:18px;width:76px;height:92px;border:12px solid var(--accent);opacity:.85}
.side{position:absolute;right:8px;top:18px;width:82px;height:78px;background:var(--bg);border:1px solid var(--c8);opacity:.92}
.btm{position:absolute;right:8px;bottom:8px;width:170px;height:84px;background:var(--bg);border:1px solid var(--accent);opacity:.92}
.txt{position:absolute;color:var(--fg);font-family:monospace;font-size:10px;line-height:1.35;white-space:pre}
.accent{color:var(--accent)}
.palette{margin-top:8px}
.swatch{display:inline-block;width:20px;height:12px;margin:1px;border-radius:2px;border:1px solid #444}
.group{margin:8px 0;padding:10px;border:1px solid #3b4261;border-radius:6px;background:#1e2030}
.group .gname{font-weight:bold;color:#a9b1d6}
.group select{margin-left:8px;padding:3px;background:#24283b;color:#c0caf5;border:1px solid #3b4261}
.override{margin:4px 0 4px 20px;font-size:13px}
.override select{margin-left:6px;padding:2px;background:#24283b;color:#c0caf5;border:1px solid #3b4261;font-size:12px}
button{padding:8px 16px;margin:4px;background:#7aa2f7;color:#1a1b26;border:none;border-radius:4px;cursor:pointer;font-weight:bold}
button:hover{background:#82aaff}
button.small{padding:3px 8px;font-size:12px}
#status{margin-top:10px;padding:6px;border-radius:4px}
#status.error{background:#4a1a1a;color:#ff6b6b}
#status.ok{background:#1a3a1a;color:#9ece6a}
</style></head><body>
<h1>omarchy-themegen Preview</h1>
<div class="info">Local only • Session token required • Changes sync to TUI • previews: %s</div>
<div><button onclick="setMode('whole-theme')">Whole Theme</button><button onclick="setMode('component-mix')">Component Mix</button></div>
<div class="current-preview"><h3>Current Preview</h3><img id="currentPreview" alt="Current selection preview"></div>
<h3>Theme Directions</h3>
<div id="directions">Loading...</div>
<h3 id="mixTitle" style="display:none">Component Mix</h3>
<div id="groups"></div>
<div id="overridesTitle" style="display:none"><h3>Per-Surface Overrides</h3></div>
<div id="overrides"></div>
<div id="status" class="ok"></div>
<script>
const token="%s"
const directions=%s
const api=(url,body)=>{fetch(url,{method:"POST",body}).then(r=>{if(!r.ok)status(r.status+": "+url,"error");else loadState()})}
const get=(url)=>{return fetch(url).then(r=>r.json())}
function status(m,c){var s=document.getElementById("status");s.textContent=m;s.className=c||"ok"}
function selectDirection(id){api("/api/select-direction?id="+id+"&token="+token)}
function selectGroup(gid,id){api("/api/select-group?group="+gid+"&id="+id+"&token="+token)}
function setMode(mode){api("/api/set-mode?mode="+mode+"&token="+token)}
function setOverride(surf,id){if(id=="0")api("/api/clear-override?surface="+surf+"&token="+token);else api("/api/set-override?surface="+surf+"&id="+id+"&token="+token)}

const groups=[{id:"desktop-shell",name:"Desktop Shell",surfaces:["waybar","hyprland","hyprlock","mako","walker","swayosd"]},
{id:"terminals-and-tui",name:"Terminals And TUI",surfaces:["ghostty","alacritty","foot","kitty","btop","terminal-palette"]},
{id:"editor",name:"Editor",surfaces:["neovim"]},
{id:"assets-and-system",name:"Assets And System",surfaces:["wallpaper-background","preview-assets","icons","light-mode","chromium","keyboard-rgb"]}]
const surfaces=["waybar","hyprland","hyprlock","mako","walker","swayosd","ghostty","alacritty","foot","kitty","btop","terminal-palette","neovim","wallpaper-background","preview-assets","icons","light-mode","chromium","keyboard-rgb"]

function renderDirections(dirs){
var h=""
dirs.forEach(d=>{
h+='<div class="dir '+(window.currentSelected==d.ID?'selected':'')+'" onclick="selectDirection('+d.ID+')"><div class="label">Direction '+d.ID+': '+d.Label+'</div>'
if(d.Colors){
var c=d.Colors
h+='<img class="preview-img" src="/preview/direction/'+d.ID+'?token='+token+'" alt="Direction '+d.ID+' preview" onerror="this.style.display=\'none\';this.nextElementSibling.style.display=\'block\'">'
h+='<div class="mock" style="--bg:'+c.Background+';--fg:'+c.Foreground+';--accent:'+c.Accent+';--c8:'+c.Color8+'"><div class="bar"></div><div class="chrome"></div><div class="term"></div><div class="logo"></div><div class="side"></div><div class="btm"></div><div class="txt" style="left:18px;top:126px">~ > ls -l\n<span class="accent">drwx Desktop\ndrwx Downloads\ndrwx Projects</span></div><div class="txt" style="right:18px;top:32px">Hardware\nCPU GPU RAM\n<span class="accent">Omarchy</span></div><div class="txt" style="right:18px;bottom:20px">cpu mem disk\n<span class="accent">██████░░░</span></div></div>'
h+='<div class="palette"><span class="swatch" style="background:'+c.Background+'" title="bg"></span><span class="swatch" style="background:'+c.Foreground+'" title="fg"></span><span class="swatch" style="background:'+c.Accent+'" title="accent"></span><span class="swatch" style="background:'+c.Color1+'"></span><span class="swatch" style="background:'+c.Color2+'"></span><span class="swatch" style="background:'+c.Color3+'"></span><span class="swatch" style="background:'+c.Color4+'"></span><span class="swatch" style="background:'+c.Color5+'"></span><span class="swatch" style="background:'+c.Color6+'"></span></div>'
}
h+='</div>'
})
document.getElementById("directions").innerHTML=h
}

function renderGroups(state){
var h=""
groups.forEach(g=>{var selected=(state.groups&&state.groups[g.id])||state.selected||1;h+='<div class="group"><span class="gname">'+g.name+'</span><select onchange="selectGroup(\''+g.id+'\',this.value)"><option value="">-- select --</option><option value="1" '+(selected==1?'selected':'')+'>Direction 1</option><option value="2" '+(selected==2?'selected':'')+'>Direction 2</option><option value="3" '+(selected==3?'selected':'')+'>Direction 3</option></select></div>'})
document.getElementById("groups").innerHTML=h
document.getElementById("mixTitle").style.display=state.mode=="component-mix"?"block":"none"
document.getElementById("groups").style.display=state.mode=="component-mix"?"block":"none"
}

function renderOverrides(state){
var h=""
surfaces.forEach(s=>{
var selected=(state.overrides&&state.overrides[s])||0
h+='<div class="override">'+s+': <select onchange="setOverride(\''+s+'\',this.value)"><option value="0" '+(selected==0?'selected':'')+'>none</option><option value="1" '+(selected==1?'selected':'')+'>Direction 1</option><option value="2" '+(selected==2?'selected':'')+'>Direction 2</option><option value="3" '+(selected==3?'selected':'')+'>Direction 3</option></select></div>'
})
document.getElementById("overrides").innerHTML=h
document.getElementById("overridesTitle").style.display=state.mode=="component-mix"?"block":"none"
}

function renderCurrentPreview(){
var img=document.getElementById("currentPreview")
img.src="/preview/current?token="+token+"&v="+Date.now()
}

function loadState(){
get("/api/state?token="+token).then(s=>{
window.currentSelected=s.selected
renderDirections(directions)
renderGroups(s)
renderOverrides(s)
renderCurrentPreview()
status("mode: "+s.mode+" | selected: "+s.selected,"ok")
}).catch(e=>status("Cannot load state","error"))
}
loadState()
</script></body></html>`
