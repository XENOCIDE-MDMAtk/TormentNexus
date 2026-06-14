package tools

/**
 * @file registry.go
 * @module go/internal/tools
 *
 * WHAT: Go-native registry for standard library and parity tools.
 * Maps tool names to their native Go implementations.
 */

import (
	"context"
	"fmt"
	"sync"
)

type ToolHandler func(ctx context.Context, args map[string]interface{}) (ToolResponse, error)

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]ToolHandler
}

func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[string]ToolHandler),
	}
	r.registerAll()
	return r
}

func (r *Registry) registerAll() {
	// Native Handlers
	r.handlers["read_file"] = HandleRead
	r.handlers["write_file"] = HandleWrite
	r.handlers["edit_file"] = HandleEdit
	r.handlers["str_replace_editor"] = HandleEdit
	r.handlers["grep_search"] = HandleGrep
	r.handlers["search_files"] = HandleGrep
	r.handlers["glob"] = HandleGlob
	r.handlers["find_files"] = HandleGlob
	r.handlers["apply_patch"] = HandleApplyPatch
	r.handlers["multi_edit"] = HandleMultiEdit
	r.handlers["bash"] = HandleBash
	r.handlers["ls"] = HandleLS
	r.handlers["list_directory"] = HandleLS
	r.handlers["web_fetch"] = HandleWebFetch
// TODO: 	r.handlers["ingest_git"] = HandleGitIngest
// TODO: 	r.handlers["sqlite_get_catalog"] = HandleSqliteGetCatalog
// TODO: 	r.handlers["sqlite_execute"] = HandleSqliteExecute
	r.handlers["search"] = HandleDDGSearch
	r.handlers["fetch_content"] = HandleDDGFetchContent
// TODO: 	r.handlers["slack_list_channels"] = HandleSlackListChannels
// TODO: 	r.handlers["slack_post_message"] = HandleSlackPostMessage
// TODO: 	r.handlers["slack_reply_to_thread"] = HandleSlackReplyToThread
// TODO: 	r.handlers["slack_add_reaction"] = HandleSlackAddReaction
// TODO: 	r.handlers["slack_get_channel_history"] = HandleSlackGetChannelHistory
// TODO: 	r.handlers["slack_get_thread_replies"] = HandleSlackGetThreadReplies
// TODO: 	r.handlers["slack_get_users"] = HandleSlackGetUsers
// TODO: 	r.handlers["slack_get_user_profile"] = HandleSlackGetUserProfile

	// Filesystem MCP Tools
// TODO: 	r.handlers["read_text_file"] = HandleReadTextFile
// TODO: 	r.handlers["create_directory"] = HandleCreateDirectory
// TODO: 	r.handlers["list_directory"] = HandleListDirectory
// TODO: 	r.handlers["list_directory_with_sizes"] = HandleListDirectoryWithSizes
// TODO: 	r.handlers["directory_tree"] = HandleDirectoryTree
// TODO: 	r.handlers["move_file"] = HandleMoveFile
// TODO: 	r.handlers["get_file_info"] = HandleGetFileInfo
// TODO: 	r.handlers["search_files"] = HandleSearchFiles

	// Ollama MCP Tools (AI & LLM Integration)
// TODO: 	r.handlers["list_local_models"] = HandleListLocalModels
// TODO: 	r.handlers["local_llm_chat"] = HandleLocalLLMChat
// TODO: 	r.handlers["ollama_health_check"] = HandleOllamaHealthCheck
// TODO: 	r.handlers["system_resource_check"] = HandleSystemResourceCheck

	// TTS MCP Tools (Media & Design)
// TODO: 	r.handlers["say_tts"] = HandleSayTTS
// TODO: 	r.handlers["openai_tts"] = HandleOpenAITTS

	// Vercel MCP Tools (Cloud & DevOps)
// TODO: 	r.handlers["vercel_list_projects"] = HandleVercelListProjects
// TODO: 	r.handlers["vercel_get_project"] = HandleVercelGetProject
// TODO: 	r.handlers["vercel_list_deployments"] = HandleVercelListDeployments
// TODO: 	r.handlers["vercel_get_deployment"] = HandleVercelGetDeployment
// TODO: 	r.handlers["vercel_cancel_deployment"] = HandleVercelCancelDeployment
// TODO: 	r.handlers["vercel_list_env_vars"] = HandleVercelListEnvVars
// TODO: 	r.handlers["vercel_create_env_var"] = HandleVercelCreateEnvVar
// TODO: 	r.handlers["vercel_delete_env_var"] = HandleVercelDeleteEnvVar

	// DexPaprika MCP Tools (Finance & Crypto)
// TODO: 	r.handlers["getCapabilities"] = HandleDexPaprikaGetCapabilities
// TODO: 	r.handlers["getNetworks"] = HandleDexPaprikaGetNetworks
// TODO: 	r.handlers["getStats"] = HandleDexPaprikaGetStats
// TODO: 	r.handlers["search"] = HandleDexPaprikaSearch
// TODO: 	r.handlers["getNetworkDexes"] = HandleDexPaprikaGetNetworkDexes
// TODO: 	r.handlers["getNetworkPools"] = HandleDexPaprikaGetNetworkPools
// TODO: 	r.handlers["getDexPools"] = HandleDexPaprikaGetDexPools
// TODO: 	r.handlers["getNetworkPoolsFilter"] = HandleDexPaprikaGetNetworkPoolsFilter
// TODO: 	r.handlers["getPoolDetails"] = HandleDexPaprikaGetPoolDetails
// TODO: 	r.handlers["getPoolOHLCV"] = HandleDexPaprikaGetPoolOHLCV
// TODO: 	r.handlers["getPoolTransactions"] = HandleDexPaprikaGetPoolTransactions
// TODO: 	r.handlers["getTokenDetails"] = HandleDexPaprikaGetTokenDetails
// TODO: 	r.handlers["getTokenPools"] = HandleDexPaprikaGetTokenPools
// TODO: 	r.handlers["getTokenMultiPrices"] = HandleDexPaprikaGetTokenMultiPrices
// TODO: 	r.handlers["filterNetworkTokens"] = HandleDexPaprikaFilterNetworkTokens
// TODO: 	r.handlers["getTopTokens"] = HandleDexPaprikaGetTopTokens
// TODO: 	r.handlers["submitFeedback"] = HandleDexPaprikaSubmitFeedback

	// National Weather Service (NWS) MCP Tools (Weather & Location)
// TODO: 	r.handlers["nws_get_forecast"] = HandleNWSGetForecast
// TODO: 	r.handlers["nws_search_alerts"] = HandleNWSSearchAlerts
// TODO: 	r.handlers["nws_get_observations"] = HandleNWSGetObservations
// TODO: 	r.handlers["nws_find_stations"] = HandleNWSFindStations
// TODO: 	r.handlers["nws_list_alert_types"] = HandleNWSListAlertTypes
// TODO: 	r.handlers["nws_get_office_discussion"] = HandleNWSGetOfficeDiscussion
// TODO: 	r.handlers["nws_get_zone_forecast"] = HandleNWSGetZoneForecast

	// ast-grep-mcp Tools (Category 11)
// TODO: 	r.handlers["ast_grep_dump_syntax_tree"] = HandleDumpSyntaxTree
// TODO: 	r.handlers["ast_grep_test_match_code_rule"] = HandleTestMatchCodeRule
// TODO: 	r.handlers["ast_grep_find_code"] = HandleFindCode
// TODO: 	r.handlers["ast_grep_find_code_by_rule"] = HandleFindCodeByRule

	// PAL Tools (Category 12)
// TODO: 	r.handlers["pal_chat"] = HandlePalChat
// TODO: 	r.handlers["pal_thinkdeep"] = HandlePalThinkDeep
// TODO: 	r.handlers["pal_planner"] = HandlePalPlanner
// TODO: 	r.handlers["pal_consensus"] = HandlePalConsensus
// TODO: 	r.handlers["pal_codereview"] = HandlePalCodeReview
// TODO: 	r.handlers["pal_precommit"] = HandlePalPrecommit
// TODO: 	r.handlers["pal_debug"] = HandlePalDebug
// TODO: 	r.handlers["pal_challenge"] = HandlePalChallenge

	// Short/alias mappings for PAL tools without prefix
// TODO: 	r.handlers["chat"] = HandlePalChat
// TODO: 	r.handlers["thinkdeep"] = HandlePalThinkDeep
// TODO: 	r.handlers["planner"] = HandlePalPlanner
// TODO: 	r.handlers["consensus"] = HandlePalConsensus
// TODO: 	r.handlers["codereview"] = HandlePalCodeReview
// TODO: 	r.handlers["precommit"] = HandlePalPrecommit
// TODO: 	r.handlers["debug"] = HandlePalDebug
// TODO: 	r.handlers["challenge"] = HandlePalChallenge

	// Serena Tools (Category 13)
// TODO: 	r.handlers["get_symbols_overview"] = HandleGetSymbolsOverview
// TODO: 	r.handlers["find_symbol"] = HandleFindSymbol
// TODO: 	r.handlers["find_referencing_symbols"] = HandleFindReferencingSymbols
// TODO: 	r.handlers["find_implementations"] = HandleFindImplementations
// TODO: 	r.handlers["find_declaration"] = HandleFindDeclaration
// TODO: 	r.handlers["rename_symbol"] = HandleRenameSymbol
// TODO: 	r.handlers["onboarding"] = HandleOnboarding

	// Claude Code Aliases
	r.handlers["Read"] = HandleRead
	r.handlers["Write"] = HandleWrite
	r.handlers["Edit"] = HandleEdit
	r.handlers["Bash"] = HandleBash
	r.handlers["LS"] = HandleLS
	r.handlers["WebFetch"] = HandleWebFetch
	r.handlers["Glob"] = HandleGlob
	r.handlers["Grep"] = HandleGrep
	r.handlers["MultiEdit"] = HandleMultiEdit

	// Codex Aliases
	r.handlers["shell"] = HandleBash
	r.handlers["create_file"] = HandleWrite
	r.handlers["view_file"] = HandleRead
	r.handlers["apply_diff"] = HandleApplyPatch
	r.handlers["search_files_codex"] = HandleGrep

	// OpenCode / Pi Aliases
	r.handlers["read"] = HandleRead
	r.handlers["write"] = HandleWrite
	r.handlers["edit"] = HandleEdit
	r.handlers["grep"] = HandleGrep
	r.handlers["ls"] = HandleLS
	r.handlers["glob_pi"] = HandleGlob

	// Thoughtbox Tools (Category 14)
// TODO: 	r.handlers["thoughtbox_search"] = HandleThoughtboxSearch
// TODO: 	r.handlers["thoughtbox_execute"] = HandleThoughtboxExecute
// TODO: 	r.handlers["thoughtbox_peer_notebook"] = HandleThoughtboxPeerNotebook

	// Fetch Tool (Assimilated)
// TODO: 	r.handlers["fetch"] = HandleFetch

	// Tavily Tools (Assimilated)
// TODO: 	r.handlers["tavily-search"] = HandleTavilySearch

	// Chrome DevTools Tools (Assimilated)
// TODO: 	r.handlers["chrome-devtools"] = HandleChromeDevTools

	// Firecrawl Tools (Assimilated from firecrawl-mcp)
// TODO: 	r.handlers["firecrawl_scrape"] = HandleFirecrawl
// TODO: 	r.handlers["firecrawl_crawl"] = HandleFirecrawl
// TODO: 	r.handlers["firecrawl"] = HandleFirecrawl

	// Exa Search Tools (Assimilated from SSE exa)
// TODO: 	r.handlers["exa_search"] = HandleExaSearch
// TODO: 	r.handlers["exa_find_similar"] = HandleExaFindSimilar
// TODO: 	r.handlers["exa_get_contents"] = HandleExaGetContents

	// arXiv Tools (Assimilated from arxiv-mcp-server)
// TODO: 	r.handlers["arxiv_search"] = HandleArxivSearch
// TODO: 	r.handlers["arxiv_get_paper"] = HandleArxivGetPaper
// TODO: 	r.handlers["arxiv_list_recent"] = HandleArxivListRecent

	// Semantic Scholar Tools (Assimilated from paper_search_server)
// TODO: 	r.handlers["paper_search"] = HandleSemanticScholarSearch
// TODO: 	r.handlers["paper_details"] = HandleSemanticScholarGetPaper
// TODO: 	r.handlers["paper_citations"] = HandleSemanticScholarGetCitations
// TODO: 	r.handlers["semantic_scholar_search"] = HandleSemanticScholarSearch
// TODO: 	r.handlers["semantic_scholar_paper"] = HandleSemanticScholarGetPaper

	// mem0 Memory Tools (Assimilated from @mem0/mcp-server)
// TODO: 	r.handlers["mem0_add_memory"] = HandleMem0AddMemory
// TODO: 	r.handlers["mem0_search_memory"] = HandleMem0SearchMemory
// TODO: 	r.handlers["mem0_get_memories"] = HandleMem0GetMemories
// TODO: 	r.handlers["mem0_delete_memory"] = HandleMem0DeleteMemory
// TODO: 	r.handlers["mem0_update_memory"] = HandleMem0UpdateMemory
// TODO: 	r.handlers["add_memory"] = HandleMem0AddMemory
// TODO: 	r.handlers["search_memory"] = HandleMem0SearchMemory

	// Alpaca Trading Tools (Assimilated from alpaca-mcp-server)
// TODO: 	r.handlers["alpaca_get_account"] = HandleAlpacaGetAccount
// TODO: 	r.handlers["alpaca_get_positions"] = HandleAlpacaGetPositions
// TODO: 	r.handlers["alpaca_get_orders"] = HandleAlpacaGetOrders
// TODO: 	r.handlers["alpaca_place_order"] = HandleAlpacaPlaceOrder
// TODO: 	r.handlers["alpaca_cancel_order"] = HandleAlpacaCancelOrder
// TODO: 	r.handlers["alpaca_get_bars"] = HandleAlpacaGetBars
// TODO: 	r.handlers["alpaca_get_latest_quote"] = HandleAlpacaGetLatestQuote

	// Alpha Vantage Financial Tools (Assimilated from av-mcp)
// TODO: 	r.handlers["av_quote"] = HandleAVGlobalQuote
// TODO: 	r.handlers["av_time_series"] = HandleAVTimeSeries
// TODO: 	r.handlers["av_forex_rate"] = HandleAVForexRate
// TODO: 	r.handlers["av_crypto_rate"] = HandleAVCryptoRate
// TODO: 	r.handlers["av_symbol_search"] = HandleAVSearch
// TODO: 	r.handlers["av_economic_indicator"] = HandleAVEconomicIndicator
// TODO: 	r.handlers["alpha_vantage_quote"] = HandleAVGlobalQuote

	// Hugging Face Hub Tools (Assimilated from SSE huggingface)
// TODO: 	r.handlers["hf_search_models"] = HandleHFSearchModels
// TODO: 	r.handlers["hf_get_model"] = HandleHFGetModel
// TODO: 	r.handlers["hf_search_datasets"] = HandleHFSearchDatasets
// TODO: 	r.handlers["hf_text_generation"] = HandleHFTextGeneration
// TODO: 	r.handlers["hf_classify_text"] = HandleHFClassification
// TODO: 	r.handlers["hf_embeddings"] = HandleHFEmbeddings
// TODO: 	r.handlers["hf_search_spaces"] = HandleHFSearchSpaces

	// Semgrep Security Tools (Assimilated from semgrep + semgrepstream)
// TODO: 	r.handlers["semgrep_scan"] = HandleSemgrepScan
// TODO: 	r.handlers["semgrep_cloud_scan"] = HandleSemgrepCloudScan
// TODO: 	r.handlers["semgrep_search_rules"] = HandleSemgrepRuleSearch

	// Octagon Financial Intelligence (Assimilated from octagon + octagon-deep-research)
// TODO: 	r.handlers["octagon_research"] = HandleOctagonResearch
// TODO: 	r.handlers["octagon_company_search"] = HandleOctagonCompanySearch
// TODO: 	r.handlers["octagon_financials"] = HandleOctagonFinancials
// TODO: 	r.handlers["octagon_news"] = HandleOctagonNews

	// Browser Automation Tools (Assimilated from playwright/browser-use/browsermcp/puppeteer/browserbase)
r.handlers["browser_navigate"] = HandleBrowserNavigate
r.handlers["browser_screenshot"] = HandleBrowserScreenshot
	r.handlers["browser_get_html"] = HandleBrowserGetHTML
	r.handlers["browser_evaluate"] = HandleBrowserEvaluate
	r.handlers["browser_click"] = HandleBrowserClick
	r.handlers["browser_fill_form"] = HandleBrowserFillForm

	// ChromaDB Vector Store Tools (Assimilated from chroma-mcp)
// TODO: 	r.handlers["chroma_list_collections"] = HandleChromaListCollections
// TODO: 	r.handlers["chroma_create_collection"] = HandleChromaCreateCollection
// TODO: 	r.handlers["chroma_add_documents"] = HandleChromaAddDocuments
// TODO: 	r.handlers["chroma_query"] = HandleChromaQuery
// TODO: 	r.handlers["chroma_delete_collection"] = HandleChromaDeleteCollection
// TODO: 	r.handlers["chroma_get_documents"] = HandleChromaGetCollection

	// Basic Memory Tools (Assimilated from basic-memory)
// TODO: 	r.handlers["basic_memory_write"] = HandleBasicMemoryWrite
// TODO: 	r.handlers["basic_memory_read"] = HandleBasicMemoryRead
// TODO: 	r.handlers["basic_memory_search"] = HandleBasicMemorySearch
// TODO: 	r.handlers["basic_memory_list"] = HandleBasicMemoryList
// TODO: 	r.handlers["basic_memory_delete"] = HandleBasicMemoryDelete
// TODO: 	r.handlers["memory_write"] = HandleBasicMemoryWrite
// TODO: 	r.handlers["memory_read"] = HandleBasicMemoryRead
// TODO: 	r.handlers["memory_search"] = HandleBasicMemorySearch

	// MindsDB ML Database Tools (Assimilated from SSE mindsdb)
// TODO: 	r.handlers["mindsdb_query"] = HandleMindsDBQuery
// TODO: 	r.handlers["mindsdb_list_models"] = HandleMindsDBListModels
// TODO: 	r.handlers["mindsdb_predict"] = HandleMindsDBPredict

	// ═══════════════════════════════════════════════════════════════
	// ASSIMILATED MCP SERVERS — Phase 2: Full Native Reimplementation
	// ═══════════════════════════════════════════════════════════════

	// GitHub Copilot API Tools (Assimilated from github SSE)// TODO: 
// TODO: 	r.handlers["github_list_repos"] = HandleGithubListRepos
// TODO: 	r.handlers["github_get_repo"] = HandleGithubGetRepo
// TODO: 	r.handlers["github_create_issue"] = HandleGithubCreateIssue
// TODO: 	r.handlers["github_list_issues"] = HandleGithubListIssues
// TODO: 	r.handlers["github_create_pr"] = HandleGithubCreatePR
// TODO: 	r.handlers["github_code_search"] = HandleGithubCodeSearch
// TODO: 	r.handlers["github_get_file_contents"] = HandleGithubGetFileContents
// TODO: 	r.handlers["github_create_or_update_file"] = HandleGithubCreateOrUpdateFile
// TODO: 	r.handlers["github_list_branches"] = HandleGithubListBranches
// TODO: 	r.handlers["github_list_workflows"] = HandleGithubListWorkflows
// TODO: 	r.handlers["github_trigger_workflow"] = HandleGithubTriggerWorkflow
// TODO: 	r.handlers["github_copilot_chat"] = HandleGithubCopilotChat

	// Supabase Tools (Assimilated from supabase SSE)
// TODO: 	r.handlers["supabase_list_projects"] = HandleSupabaseListProjects
// TODO: 	r.handlers["supabase_get_project"] = HandleSupabaseGetProject
// TODO: 	r.handlers["supabase_execute_sql"] = HandleSupabaseExecuteSQL
// TODO: 	r.handlers["supabase_select_rows"] = HandleSupabaseSelectRows
// TODO: 	r.handlers["supabase_insert_rows"] = HandleSupabaseInsertRows
// TODO: 	r.handlers["supabase_update_rows"] = HandleSupabaseUpdateRows
// TODO: 	r.handlers["supabase_delete_rows"] = HandleSupabaseDeleteRows
// TODO: 	r.handlers["supabase_list_tables"] = HandleSupabaseListTables
// TODO: 	r.handlers["supabase_invoke_function"] = HandleSupabaseInvokeFunction

	// Desktop Commander Tools (Assimilated from @wonderwhy-er/desktop-commander)
// TODO: 	r.handlers["desktop_execute_command"] = HandleDesktopExecuteCommand
// TODO: 	r.handlers["desktop_read_file"] = HandleDesktopReadFile
// TODO: 	r.handlers["desktop_read_multiple_files"] = HandleDesktopReadMultipleFiles
// TODO: 	r.handlers["desktop_write_file"] = HandleDesktopWriteFile
// TODO: 	r.handlers["desktop_create_directory"] = HandleDesktopCreateDirectory
// TODO: 	r.handlers["desktop_list_directory"] = HandleDesktopListDirectory
// TODO: 	r.handlers["desktop_directory_tree"] = HandleDesktopDirectoryTree
// TODO: 	r.handlers["desktop_search_files"] = HandleDesktopSearchFiles
// TODO: 	r.handlers["desktop_move_file"] = HandleDesktopMoveFile
// TODO: 	r.handlers["desktop_get_file_info"] = HandleDesktopGetFileInfo
// TODO: 	r.handlers["desktop_list_processes"] = HandleDesktopListProcesses
// TODO: 	r.handlers["desktop_kill_process"] = HandleDesktopKillProcess
// TODO: 	r.handlers["desktop_get_system_info"] = HandleDesktopGetSystemInfo
// TODO: 	r.handlers["desktop_execute_script"] = HandleDesktopExecuteScript
// TODO: 	r.handlers["desktop_open_file"] = HandleDesktopOpenFile
// TODO: 	r.handlers["desktop_tail_file"] = HandleDesktopTailFile

	// Gemini API Tools (Assimilated from gemini-mcp)
// TODO: 	r.handlers["gemini_chat"] = HandleGeminiChat
// TODO: 	r.handlers["gemini_code_generation"] = HandleGeminiCodeGeneration
// TODO: 	r.handlers["gemini_vision"] = HandleGeminiVision
// TODO: 	r.handlers["gemini_embeddings"] = HandleGeminiEmbeddings
// TODO: 	r.handlers["gemini_list_models"] = HandleGeminiListModels
// TODO: 	r.handlers["gemini_function_calling"] = HandleGeminiFunctionCalling

	// DBHub Universal Database Tools (Assimilated from @bytebase/dbhub)
// TODO: 	r.handlers["dbhub_list_databases"] = HandleDBHubListDatabases
// TODO: 	r.handlers["dbhub_list_tables"] = HandleDBHubListTables
// TODO: 	r.handlers["dbhub_describe_table"] = HandleDBHubDescribeTable
// TODO: 	r.handlers["dbhub_execute_query"] = HandleDBHubExecuteQuery
// TODO: 	r.handlers["dbhub_list_schemas"] = HandleDBHubListSchemas

	// ConPort Context Portal Tools (Assimilated from context-portal-mcp)
// TODO: 	r.handlers["conport_get_context"] = HandleConPortGetContext
// TODO: 	r.handlers["conport_update_context"] = HandleConPortUpdateContext
// TODO: 	r.handlers["conport_log_decision"] = HandleConPortLogDecision
// TODO: 	r.handlers["conport_get_decisions"] = HandleConPortGetDecisions
// TODO: 	r.handlers["conport_add_pattern"] = HandleConPortAddPattern
// TODO: 	r.handlers["conport_get_patterns"] = HandleConPortGetPatterns
// TODO: 	r.handlers["conport_set_active_context"] = HandleConPortSetActiveContext
// TODO: 	r.handlers["conport_get_active_context"] = HandleConPortGetActiveContext
// TODO: 	r.handlers["conport_log_progress"] = HandleConPortLogProgress
// TODO: 	r.handlers["conport_get_progress"] = HandleConPortGetProgress

	// ChunkHound Code Search Tools (Assimilated from chunkhound)
// TODO: 	r.handlers["chunkhound_index"] = HandleChunkhoundIndex
// TODO: 	r.handlers["chunkhound_search"] = HandleChunkhoundSearch
// TODO: 	r.handlers["chunkhound_stats"] = HandleChunkhoundStats
// TODO: 	r.handlers["chunkhound_list_indexed"] = HandleChunkhoundListIndexed
// TODO: 	r.handlers["chunkhound_get_chunk"] = HandleChunkhoundGetChunk

	// NotebookLM Tools (Assimilated from @roomi-fields/notebooklm-mcp)
// TODO: 	r.handlers["notebooklm_create_notebook"] = HandleNotebookLMCreateNotebook
// TODO: 	r.handlers["notebooklm_query_notebook"] = HandleNotebookLMQueryNotebook
// TODO: 	r.handlers["notebooklm_list_notebooks"] = HandleNotebookLMListNotebooks
// TODO: 	r.handlers["notebooklm_add_source"] = HandleNotebookLMAddSource
// TODO: 	r.handlers["notebooklm_get_summary"] = HandleNotebookLMGetSummary
// TODO: 	r.handlers["notebooklm_upload_pdf"] = HandleNotebookLMUploadPDF

	// Vibe Check Tools (Assimilated from @pv-bhat/vibe-check-mcp)
// TODO: 	r.handlers["vibe_check_analyze"] = HandleVibeCheckAnalyze
// TODO: 	r.handlers["vibe_check_quick"] = HandleVibeCheckQuick
// TODO: 	r.handlers["vibe_check_review_patterns"] = HandleVibeCheckReviewPatterns

	// SuperMemory Tools (Assimilated from mcp-supermemory-ai)
// TODO: 	r.handlers["supermemory_add"] = HandleSuperMemoryAdd
// TODO: 	r.handlers["supermemory_search"] = HandleSuperMemorySearch
// TODO: 	r.handlers["supermemory_delete"] = HandleSuperMemoryDelete
// TODO: 	r.handlers["supermemory_list"] = HandleSuperMemoryList

	// Probe Code Search Tools (Assimilated from @probelabs/probe)
// TODO: 	r.handlers["probe_search_code"] = HandleProbeSearchCode
// TODO: 	r.handlers["probe_find_symbol"] = HandleProbeFindSymbol
// TODO: 	r.handlers["probe_get_structure"] = HandleProbeGetStructure
// TODO: 	r.handlers["probe_explain_code"] = HandleProbeExplainCode

	// Cipher Memory Aggregator Tools (Assimilated from @byterover/cipher)
// TODO: 	r.handlers["cipher_add_memory"] = HandleCipherAddMemory
// TODO: 	r.handlers["cipher_search_memory"] = HandleCipherSearchMemory
// TODO: 	r.handlers["cipher_list_memories"] = HandleCipherListMemories
// TODO: 	r.handlers["cipher_delete_memory"] = HandleCipherDeleteMemory
// TODO: 	r.handlers["cipher_ask"] = HandleCipherAskCipher

	// DeepContext Code Understanding Tools (Assimilated from @wildcard-ai/deepcontext)
// TODO: 	r.handlers["deepcontext_analyze"] = HandleDeepContextAnalyzeCodebase
// TODO: 	r.handlers["deepcontext_get_context"] = HandleDeepContextGetContext
// TODO: 	r.handlers["deepcontext_find_patterns"] = HandleDeepContextFindPatterns
// TODO: 	r.handlers["deepcontext_summarize_architecture"] = HandleDeepContextSummarizeArchitecture

	// Windows MCP Tools (Assimilated from windows-mcp)
// TODO: 	r.handlers["windows_get_system_info"] = HandleWindowsMCPGetSystemInfo
// TODO: 	r.handlers["windows_list_services"] = HandleWindowsMCPListServices
// TODO: 	r.handlers["windows_get_service"] = HandleWindowsMCPGetService
// TODO: 	r.handlers["windows_list_processes"] = HandleWindowsMCPListProcesses
// TODO: 	r.handlers["windows_read_registry"] = HandleWindowsMCPReadRegistry
// TODO: 	r.handlers["windows_open_application"] = HandleWindowsMCPOpenApplication
// TODO: 	r.handlers["windows_get_clipboard"] = HandleWindowsMCPGetClipboard
// TODO: 	r.handlers["windows_set_clipboard"] = HandleWindowsMCPSetClipboard
// TODO: 	r.handlers["windows_list_drives"] = HandleWindowsMCPListDrives
// TODO: 	r.handlers["windows_get_event_log"] = HandleWindowsMCPGetEventLog

	// Prism Code Quality Tools (Assimilated from prism-mcp-server)
// TODO: 	r.handlers["prism_analyze_quality"] = HandlePrismAnalyzeQuality
// TODO: 	r.handlers["prism_suggest_refactor"] = HandlePrismSuggestRefactor
// TODO: 	r.handlers["prism_detect_smells"] = HandlePrismDetectSmells
// TODO: 	r.handlers["prism_transform_code"] = HandlePrismTransformCode

	// TaskMaster AI Task Management Tools (Assimilated from task-master-ai)
// TODO: 	r.handlers["taskmaster_create_task"] = HandleTaskMasterCreateTask
// TODO: 	r.handlers["taskmaster_get_task"] = HandleTaskMasterGetTask
// TODO: 	r.handlers["taskmaster_list_tasks"] = HandleTaskMasterListTasks
// TODO: 	r.handlers["taskmaster_update_status"] = HandleTaskMasterUpdateStatus
// TODO: 	r.handlers["taskmaster_add_subtask"] = HandleTaskMasterAddSubtask
// TODO: 	r.handlers["taskmaster_next_task"] = HandleTaskMasterNextTask
// TODO: 	r.handlers["taskmaster_generate_from_prd"] = HandleTaskMasterGenerateFromPRD
// TODO: 	r.handlers["taskmaster_expand_task"] = HandleTaskMasterExpandTask

	// ═══════════════════════════════════════════════════════════════
	// SKILL REGISTRY - Database-backed skill management with deduplication
	// ═══════════════════════════════════════════════════════════════

	// Skill Registry Tools// TODO: 
// TODO: 	r.handlers["skill_list"] = HandleSkillList
// TODO: 	r.handlers["skill_get"] = HandleSkillGet
// TODO: 	r.handlers["skill_store"] = HandleSkillStore
// TODO: 	r.handlers["skill_search"] = HandleSkillSearch
// TODO: 	r.handlers["skills_list"] = HandleSkillList
// TODO: 	r.handlers["skills_get"] = HandleSkillGet
// TODO: 	r.handlers["skills_store"] = HandleSkillStore
// TODO: 	r.handlers["skills_search"] = HandleSkillSearch

	// OpenMemory — local persistent memory store
// TODO: 	r.handlers["openmemory_add"] = HandleOpenMemoryAdd
// TODO: 	r.handlers["openmemory_search"] = HandleOpenMemorySearch
// TODO: 	r.handlers["openmemory_get"] = HandleOpenMemoryGet
// TODO: 	r.handlers["openmemory_delete"] = HandleOpenMemoryDelete
// TODO: 	r.handlers["openmemory_list"] = HandleOpenMemoryList

	// AutoMem — graph-vector memory for AI agents
// TODO: 	r.handlers["automem_add"] = HandleAutoMemAdd
// TODO: 	r.handlers["automem_search"] = HandleAutoMemSearch
// TODO: 	r.handlers["automem_get"] = HandleAutoMemGet
// TODO: 	r.handlers["automem_delete"] = HandleAutoMemDelete
// TODO: 	r.handlers["automem_list"] = HandleAutoMemList
// TODO: 	r.handlers["automem_associate"] = HandleAutoMemAssociate

	// lsmcp — LSP code manipulation and analysis
// TODO: 	r.handlers["project_overview"] = HandleLsmcpProjectOverview
// TODO: 	r.handlers["search_symbols"] = HandleLsmcpSearchSymbols
// TODO: 	r.handlers["get_diagnostics"] = HandleLsmcpGetDiagnostics
// TODO: 	r.handlers["find_references"] = HandleLsmcpFindReferences
// TODO: 	r.handlers["get_symbol_details"] = HandleLsmcpGetSymbolDetails

	// CodeAlive — semantic code search and context engine
// TODO: 	r.handlers["codealive_search"] = HandleCodeAliveSearch
// TODO: 	r.handlers["codealive_grep"] = HandleCodeAliveGrep
// TODO: 	r.handlers["codealive_ask"] = HandleCodeAliveAsk

	// Prometheus MCP — monitoring queries
// TODO: 	r.handlers["prom_query"] = HandlePromQuery
// TODO: 	r.handlers["prom_alerts"] = HandlePromAlerts
// TODO: 	r.handlers["prom_targets"] = HandlePromTargets
// TODO: 	r.handlers["prom_metadata"] = HandlePromMetadata

	// Smart-Thinking — graph-based reasoning
// TODO: 	r.handlers["smart_reason"] = HandleSmartReason
// TODO: 	r.handlers["smart_session"] = HandleSmartSession
// TODO: 	r.handlers["smart_evaluate"] = HandleSmartEvaluate
// TODO: 	r.handlers["smart_graph"] = HandleSmartGraph

	// Mimir — Neo4j-backed persistent memory
// TODO: 	r.handlers["mimir_store"] = HandleMimirStore
// TODO: 	r.handlers["mimir_search"] = HandleMimirSearch
// TODO: 	r.handlers["mimir_retrieve"] = HandleMimirRetrieve
// TODO: 	r.handlers["mimir_connect"] = HandleMimirConnect
// TODO: 	r.handlers["mimir_forget"] = HandleMimirForget

	// Sysmon — system monitoring
// TODO: 	r.handlers["sysmon_overview"] = HandleSysmonOverview
// TODO: 	r.handlers["sysmon_health"] = HandleSysmonHealth
// TODO: 	r.handlers["sysmon_top"] = HandleSysmonTop
// TODO: 	r.handlers["sysmon_disk"] = HandleSysmonDisk
// TODO: 	r.handlers["sysmon_network"] = HandleSysmonNetwork
// TODO: 	r.handlers["sysmon_find"] = HandleSysmonFind

	// Docker — container management
// TODO: 	r.handlers["docker_list_containers"] = HandleDockerListContainers
// TODO: 	r.handlers["docker_list_images"] = HandleDockerListImages
// TODO: 	r.handlers["docker_inspect"] = HandleDockerInspect
// TODO: 	r.handlers["docker_logs"] = HandleDockerLogs
// TODO: 	r.handlers["docker_stats"] = HandleDockerStats
// TODO: 	r.handlers["docker_exec"] = HandleDockerExec

	// Social — Twitter/X and Reddit
// TODO: 	r.handlers["twitter_search"] = HandleTwitterSearch
// TODO: 	r.handlers["twitter_user_timeline"] = HandleTwitterUserTimeline
// TODO: 	r.handlers["reddit_search"] = HandleRedditSearch
// TODO: 	r.handlers["reddit_get_posts"] = HandleRedditGetPosts

	// Git — repository operations
// TODO: 	r.handlers["git_status"] = HandleGitStatus
// TODO: 	r.handlers["git_log"] = HandleGitLog
// TODO: 	r.handlers["git_diff"] = HandleGitDiff
// TODO: 	r.handlers["git_branches"] = HandleGitBranches
// TODO: 	r.handlers["git_show"] = HandleGitShow
// TODO: 	r.handlers["git_blame"] = HandleGitBlame
// TODO: 	r.handlers["git_commit"] = HandleGitCommit
// TODO: 	r.handlers["git_checkout"] = HandleGitCheckout

	// Terraform — infrastructure management
// TODO: 	r.handlers["terraform_search_providers"] = HandleTerraformSearchProviders
// TODO: 	r.handlers["terraform_search_modules"] = HandleTerraformSearchModules
// TODO: 	r.handlers["terraform_get_provider"] = HandleTerraformGetProvider

	// Google News — news headlines and search
// TODO: 	r.handlers["google_news_headlines"] = HandleGoogleNewsHeadlines
// TODO: 	r.handlers["google_news_search"] = HandleGoogleNewsSearch

	// OpenRouter Deep Research
// TODO: 	r.handlers["deep_research"] = HandleDeepResearch
// TODO: 	r.handlers["deep_research_status"] = HandleDeepResearchStatus

	// Prompt Library — SQLite-backed prompt storage
// TODO: 	r.handlers["prompt_list"] = HandlePromptList
// TODO: 	r.handlers["prompt_get"] = HandlePromptGet
// TODO: 	r.handlers["prompt_search"] = HandlePromptSearch

	// Context Server — SQLite-backed context management
// TODO: 	r.handlers["context_store"] = HandleContextStore
// TODO: 	r.handlers["context_search"] = HandleContextSearch
// TODO: 	r.handlers["context_get"] = HandleContextGet
// TODO: 	r.handlers["context_delete"] = HandleContextDelete
// TODO: 	r.handlers["context_list_threads"] = HandleContextListThreads
// TODO: 	r.handlers["context_stats"] = HandleContextStats

	// WebPeel — web data extraction
// TODO: 	r.handlers["webpeel_fetch"] = HandleWebpeelFetch
// TODO: 	r.handlers["webpeel_search"] = HandleWebpeelSearch
// TODO: 	r.handlers["webpeel_extract"] = HandleWebpeelExtract

	// Omnisearch — universal search
// TODO: 	r.handlers["omnisearch_github"] = HandleOmnisearchGithub
// TODO: 	r.handlers["omnisearch_stackoverflow"] = HandleOmnisearchStackoverflow
// TODO: 	r.handlers["omnisearch_npm"] = HandleOmnisearchNpm
// TODO: 	r.handlers["omnisearch_pypi"] = HandleOmnisearchPypi
// TODO: 	r.handlers["omnisearch_web"] = HandleOmnisearchWeb

	// Grants — government grants discovery
// TODO: 	r.handlers["grants_search"] = HandleGrantsSearch
// TODO: 	r.handlers["grants_by_agency"] = HandleGrantsByAgency
// TODO: 	r.handlers["grants_by_category"] = HandleGrantsByCategory
// TODO: 	r.handlers["grants_trends"] = HandleGrantsTrends

	// Food Data Central — USDA nutrition database
// TODO: 	r.handlers["food_search"] = HandleFoodSearch
// TODO: 	r.handlers["food_get"] = HandleFoodGet
// TODO: 	r.handlers["food_list"] = HandleFoodList

	// Panther — security monitoring
// TODO: 	r.handlers["panther_query"] = HandlePantherQuery
// TODO: 	r.handlers["panther_list_detections"] = HandlePantherListDetections
// TODO: 	r.handlers["panther_get_findings"] = HandlePantherGetFindings
// TODO: 	r.handlers["panther_list_policies"] = HandlePantherListPolicies

	// Srclight — code indexing for AI agents
// TODO: 	r.handlers["srclight_index"] = HandleSrclightIndex
// TODO: 	r.handlers["srclight_search"] = HandleSrclightSearch
// TODO: 	r.handlers["srclight_status"] = HandleSrclightStatus
// TODO: 	r.handlers["srclight_list_languages"] = HandleSrclightListLanguages

	// Coolify — deployment & infrastructure management
// TODO: 	r.handlers["coolify_list_projects"] = HandleCoolifyListProjects
// TODO: 	r.handlers["coolify_create_project"] = HandleCoolifyCreateProject
// TODO: 	r.handlers["coolify_list_services"] = HandleCoolifyListServices
// TODO: 	r.handlers["coolify_deploy_service"] = HandleCoolifyDeployService
// TODO: 	r.handlers["coolify_get_logs"] = HandleCoolifyGetLogs
// TODO: 	r.handlers["coolify_list_databases"] = HandleCoolifyListDatabases

	// Harness Integrations
// TODO: 	r.handlers["launch_tabby"] = HandleTabby
// TODO: 	r.handlers["launch_warp"] = HandleWarp
// TODO: 	r.handlers["launch_hyper"] = HandleHyper
// TODO: 	r.handlers["launch_hyperharness"] = HandleHyperharness
// TODO: 	r.handlers["hermes_agent"] = HandleHermesAgent
// TODO: 	r.handlers["pi_mono"] = HandlePiMono

	// Bobbybookmarks Integration
// TODO: 	r.handlers["ripgrep_search"] = HandleRipgrep
// TODO: 	r.handlers["anyquery"] = HandleAnyquery
// TODO: 	r.handlers["codemod"] = HandleCodemod
// TODO: 	r.handlers["puppeteer_navigate"] = HandlePuppeteer
	r.handlers["bobbybookmarks_sync"] = HandleBobbyBookmarksSync
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (ToolResponse, error) {
	r.mu.RLock()
	handler, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return ToolResponse{}, fmt.Errorf("tool handler not found for: %s", name)
	}
	return handler(ctx, args)
}

func (r *Registry) HasTool(name string) bool {
	r.mu.RLock()
	_, ok := r.handlers[name]
	r.mu.RUnlock()
	return ok
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		result = append(result, name)
	}
	return result
}
