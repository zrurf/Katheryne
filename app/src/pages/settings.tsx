import { createSignal, Show, onMount, For, createResource } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { authStore } from "../stores/auth";
import { appNav } from "../services/nav";
import { themeStore } from "../stores/theme";
import {
  getSavedServers,
  getActiveServer,
  setActiveServerId,
  addServer,
  removeServer,
  generateServerId,
  parseServerUrl,
  getServerApiBase,
  type ServerConfig,
} from "../services/config";
import { checkServerHealth, api } from "../services/api";
import type {
  BotTemplateItem,
  CreateTemplateReq,
  BotInstanceItem,
  CreateInstanceReq,
  KBItem,
  CreateKBReq,
  DocItem,
} from "../services/api";
import {
  ArrowLeft,
  Server,
  Plus,
  X,
  Check,
  LogOut,
  User,
  Shield,
  Bell,
  Palette,
  Globe,
  Trash2,
  Bot,
  Cpu,
  Brain,
  Edit,
  Upload,
  RefreshCw,
  FileText,
} from "lucide-solid";

export function SettingsPage() {
  const navigate = useNavigate();
  const [servers, setServers] = createSignal<ServerConfig[]>([]);
  const [activeServer, setActiveServer] = createSignal<ServerConfig | null>(null);
  const [newServerUrl, setNewServerUrl] = createSignal("");
  const [addingServer, setAddingServer] = createSignal(false);
  const [serverError, setServerError] = createSignal("");
  const [activeTab, setActiveTab] = createSignal<"general" | "servers" | "account" | "templates" | "instances" | "community" | "rag">("general");
  const [editingName, setEditingName] = createSignal(false);
  const [newName, setNewName] = createSignal("");

  const handleAvatarUpload = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      const uploadResp = await api.oss.upload(file);
      const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
      await authStore.updateProfile(undefined, proxyPath);
    } catch {
      alert("头像上传失败，请稍后再试");
    } finally {
      input.value = "";
    }
  };

  const handleSaveName = async () => {
    const name = newName().trim();
    if (!name) return;
    try {
      await authStore.updateProfile(name, undefined);
      setEditingName(false);
    } catch {
      alert("昵称修改失败");
    }
  };

  onMount(() => {
    const saved = getSavedServers();
    setServers(saved);
    const active = getActiveServer();
    setActiveServer(active);
  });

  const handleSelectServer = (server: ServerConfig) => {
    setActiveServerId(server.id);
    setActiveServer(server);
    setServerError("");
  };

  const handleAddServer = async (e: Event) => {
    e.preventDefault();
    if (!newServerUrl().trim()) return;

    setAddingServer(true);
    setServerError("");

    try {
      const { apiUrl, wsUrl, name } = parseServerUrl(newServerUrl());
      const isHealthy = await checkServerHealth(apiUrl);
      if (!isHealthy) {
        setServerError("无法连接到该服务器，请检查地址是否正确");
        setAddingServer(false);
        return;
      }

      const server: ServerConfig = {
        id: generateServerId(),
        name,
        apiUrl,
        wsUrl,
      };
      addServer(server);
      setServers(getSavedServers());
      setActiveServerId(server.id);
      setActiveServer(server);
      setNewServerUrl("");
    } catch (err) {
      setServerError(`服务器配置失败: ${err instanceof Error ? err.message : "未知错误"}`);
    } finally {
      setAddingServer(false);
    }
  };

  const handleRemoveServer = (id: string) => {
    removeServer(id);
    const updated = getSavedServers();
    setServers(updated);
    if (activeServer()?.id === id) {
      setActiveServer(updated.length > 0 ? updated[0] : null);
      if (updated.length > 0) {
        setActiveServerId(updated[0].id);
      }
    }
  };

  const handleLogout = async () => {
    await authStore.logout();
    navigate("/login", { replace: true });
  };

  // ============ Template State ============
  const [templates, setTemplates] = createSignal<BotTemplateItem[]>([]);
  const [templatesLoading, setTemplatesLoading] = createSignal(false);
  const [showCreateTemplate, setShowCreateTemplate] = createSignal(false);
  const [editingTemplate, setEditingTemplate] = createSignal<BotTemplateItem | null>(null);
  const [tplName, setTplName] = createSignal("");
  const [tplDesc, setTplDesc] = createSignal("");
  const [tplCategory, setTplCategory] = createSignal("");
  const [tplSystemPrompt, setTplSystemPrompt] = createSignal("");
  const [tplWelcomeMsg, setTplWelcomeMsg] = createSignal("");
  const [tplTags, setTplTags] = createSignal("");
  const [tplSupportedModels, setTplSupportedModels] = createSignal("");
  const [tplTools, setTplTools] = createSignal("");
  const [tplKbStructure, setTplKbStructure] = createSignal("");
  const [tplConfigSchema, setTplConfigSchema] = createSignal("");

  const loadTemplates = async () => {
    setTemplatesLoading(true);
    try {
      const resp = await api.bot.listMyTemplates();
      setTemplates(resp.list || []);
    } catch {
      setTemplates([]);
    } finally {
      setTemplatesLoading(false);
    }
  };

  createResource(() => activeTab() === "templates", () => { if (activeTab() === "templates") loadTemplates(); });

  const handleCreateTemplate = async (e: Event) => {
    e.preventDefault();
    if (!tplName().trim()) return;
    try {
      const data: CreateTemplateReq = { name: tplName().trim() };
      if (tplDesc().trim()) data.description = tplDesc().trim();
      if (tplCategory().trim()) data.category = tplCategory().trim();
      if (tplSystemPrompt().trim()) data.system_prompt = tplSystemPrompt().trim();
      if (tplWelcomeMsg().trim()) data.welcome_message = tplWelcomeMsg().trim();
      if (tplTags().trim()) data.tags = tplTags().split(",").map(s => s.trim()).filter(Boolean);
      if (tplSupportedModels().trim()) data.supported_models = tplSupportedModels().trim();
      if (tplTools().trim()) data.tool_definitions = tplTools().trim();
      if (tplKbStructure().trim()) data.kb_structure = tplKbStructure().trim();
      if (tplConfigSchema().trim()) data.config_schema = tplConfigSchema().trim();
      await api.bot.createTemplate(data);
      resetTemplateForm();
      loadTemplates();
    } catch (err) {
      alert(`模板创建失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const resetTemplateForm = () => {
    setShowCreateTemplate(false);
    setEditingTemplate(null);
    setTplName("");
    setTplDesc("");
    setTplCategory("");
    setTplSystemPrompt("");
    setTplWelcomeMsg("");
    setTplTags("");
    setTplSupportedModels("");
    setTplTools("");
    setTplKbStructure("");
    setTplConfigSchema("");
  };

  const handleEditTemplate = async (e: Event) => {
    e.preventDefault();
    const tpl = editingTemplate();
    if (!tpl) return;
    try {
      const data: Partial<CreateTemplateReq> = {};
      if (tplName().trim()) data.name = tplName().trim();
      if (tplDesc().trim()) data.description = tplDesc().trim();
      if (tplCategory().trim()) data.category = tplCategory().trim();
      if (tplSystemPrompt().trim()) data.system_prompt = tplSystemPrompt().trim();
      if (tplWelcomeMsg().trim()) data.welcome_message = tplWelcomeMsg().trim();
      if (tplTags().trim()) data.tags = tplTags().split(",").map(s => s.trim()).filter(Boolean);
      if (tplSupportedModels().trim()) data.supported_models = tplSupportedModels().trim();
      if (tplTools().trim()) data.tool_definitions = tplTools().trim();
      if (tplKbStructure().trim()) data.kb_structure = tplKbStructure().trim();
      if (tplConfigSchema().trim()) data.config_schema = tplConfigSchema().trim();
      await api.bot.updateTemplate(tpl.template_id, data);
      resetTemplateForm();
      loadTemplates();
    } catch (err) {
      alert(`模板更新失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const startEditTemplate = (tpl: BotTemplateItem) => {
    setEditingTemplate(tpl);
    setTplName(tpl.name);
    setTplDesc(tpl.description || "");
    setTplCategory(tpl.category || "");
    setTplSystemPrompt(tpl.system_prompt || "");
    setTplWelcomeMsg(tpl.welcome_message || "");
    setTplTags((tpl.tags || []).join(", "));
    setTplSupportedModels(tpl.supported_models || "");
    setTplTools(tpl.tool_definitions || "");
    setTplKbStructure(tpl.kb_structure || "");
    setTplConfigSchema(tpl.config_schema || "");
    setShowCreateTemplate(true);
  };

  const handleDeleteTemplate = async (templateId: number) => {
    if (!confirm("确定要删除这个模板吗？关联的实例也将受到影响。")) return;
    try {
      await api.bot.deleteTemplate(templateId);
      loadTemplates();
    } catch (err) {
      alert(`删除失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  // ============ Instance State ============
  const [instances, setInstances] = createSignal<BotInstanceItem[]>([]);
  const [instancesLoading, setInstancesLoading] = createSignal(false);
  const [showCreateInstance, setShowCreateInstance] = createSignal(false);
  const [editingInstance, setEditingInstance] = createSignal<BotInstanceItem | null>(null);
  const [instTemplateId, setInstTemplateId] = createSignal("");
  const [instName, setInstName] = createSignal("");
  const [instModelProvider, setInstModelProvider] = createSignal("");
  const [instModelName, setInstModelName] = createSignal("");
  const [instApiKey, setInstApiKey] = createSignal("");
  const [instApiBaseUrl, setInstApiBaseUrl] = createSignal("");
  const [instKbConfig, setInstKbConfig] = createSignal("");
  const [instConfig, setInstConfig] = createSignal("");
  const [instSelfHosted, setInstSelfHosted] = createSignal(false);

  const loadInstances = async () => {
    setInstancesLoading(true);
    try {
      const resp = await api.bot.listMyInstances();
      setInstances(resp.list || []);
    } catch {
      setInstances([]);
    } finally {
      setInstancesLoading(false);
    }
  };

  createResource(() => activeTab() === "instances", () => { if (activeTab() === "instances") loadInstances(); });

  const handleCreateInstance = async (e: Event) => {
    e.preventDefault();
    const templateId = parseInt(instTemplateId().trim(), 10);
    if (isNaN(templateId) || templateId <= 0) {
      alert("请输入有效的模板 ID");
      return;
    }
    try {
      const data: CreateInstanceReq = { template_id: templateId };
      if (instName().trim()) data.name = instName().trim();
      if (instModelProvider().trim()) data.model_provider = instModelProvider().trim();
      if (instModelName().trim()) data.model_name = instModelName().trim();
      if (instApiKey().trim()) data.api_key = instApiKey().trim();
      if (instApiBaseUrl().trim()) data.api_base_url = instApiBaseUrl().trim();
      if (instKbConfig().trim()) data.kb_config = instKbConfig().trim();
      if (instConfig().trim()) data.instance_config = instConfig().trim();
      data.is_self_hosted = instSelfHosted();
      await api.bot.createInstance(data);
      resetInstanceForm();
      loadInstances();
    } catch (err) {
      alert(`实例创建失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const resetInstanceForm = () => {
    setShowCreateInstance(false);
    setEditingInstance(null);
    setInstTemplateId("");
    setInstName("");
    setInstModelProvider("");
    setInstModelName("");
    setInstApiKey("");
    setInstApiBaseUrl("");
    setInstKbConfig("");
    setInstConfig("");
    setInstSelfHosted(false);
  };

  const handleEditInstance = async (e: Event) => {
    e.preventDefault();
    const inst = editingInstance();
    if (!inst) return;
    try {
      const data: Partial<CreateInstanceReq> = { template_id: inst.template_id };
      if (instName().trim()) data.name = instName().trim();
      if (instModelProvider().trim()) data.model_provider = instModelProvider().trim();
      if (instModelName().trim()) data.model_name = instModelName().trim();
      if (instApiKey().trim()) data.api_key = instApiKey().trim();
      if (instApiBaseUrl().trim()) data.api_base_url = instApiBaseUrl().trim();
      if (instKbConfig().trim()) data.kb_config = instKbConfig().trim();
      if (instConfig().trim()) data.instance_config = instConfig().trim();
      data.is_self_hosted = instSelfHosted();
      await api.bot.updateInstance(inst.instance_id, data);
      resetInstanceForm();
      loadInstances();
    } catch (err) {
      alert(`实例更新失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const startEditInstance = (inst: BotInstanceItem) => {
    setEditingInstance(inst);
    setInstTemplateId(String(inst.template_id));
    setInstName(inst.name);
    setInstModelProvider(inst.model_provider || "");
    setInstModelName(inst.model_name || "");
    setInstApiKey("");
    setInstApiBaseUrl("");
    setInstKbConfig(inst.kb_config || "");
    setInstConfig("");
    setInstSelfHosted(inst.is_self_hosted);
    setShowCreateInstance(true);
  };

  const handleDeleteInstance = async (instanceId: number) => {
    if (!confirm("确定要删除这个实例吗？")) return;
    try {
      await api.bot.deleteInstance(instanceId);
      loadInstances();
    } catch (err) {
      alert(`删除失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  // ============ Community State ============
  const [communityHostedBots, setCommunityHostedBots] = createSignal<import("../services/api").CommunityHostedBot[]>([]);
  const [communityTemplates, setCommunityTemplates] = createSignal<import("../services/api").CommunityTemplate[]>([]);
  const [communityBotsLoading, setCommunityBotsLoading] = createSignal(false);
  const [communitySearch, setCommunitySearch] = createSignal("");
  const [showInstallDialog, setShowInstallDialog] = createSignal(false);
  const [installBotId, setInstallBotId] = createSignal("");
  const [convList, setConvList] = createSignal<{ conv_id: string; name: string; type: string; avatar: string }[]>([]);
  const [selectedConvIds, setSelectedConvIds] = createSignal<Set<string>>(new Set());
  const [convListLoading, setConvListLoading] = createSignal(false);

  // ============ Quick Create Instance from Community Template ============
  const [showQuickCreate, setShowQuickCreate] = createSignal(false);
  const [quickCreateTemplate, setQuickCreateTemplate] = createSignal<import("../services/api").CommunityTemplate | null>(null);
  const [quickInstName, setQuickInstName] = createSignal("");
  const [quickModelProvider, setQuickModelProvider] = createSignal("");
  const [quickModelName, setQuickModelName] = createSignal("");
  const [quickCreating, setQuickCreating] = createSignal(false);

  const openQuickCreate = (tpl: import("../services/api").CommunityTemplate) => {
    setQuickCreateTemplate(tpl);
    setQuickInstName(tpl.name);
    setQuickModelProvider("");
    setQuickModelName("");
    setShowQuickCreate(true);
  };

  const handleQuickCreate = async () => {
    const tpl = quickCreateTemplate();
    if (!tpl) return;
    setQuickCreating(true);
    try {
      // Create instance from the template
      const resp = await api.bot.createInstance({
        template_id: parseInt(tpl.template_id, 10),
        name: quickInstName().trim() || tpl.name,
        model_provider: quickModelProvider() || undefined,
        model_name: quickModelName() || undefined,
      });
      setShowQuickCreate(false);
      // Switch to instances tab and reload
      setActiveTab("instances");
      loadInstances();
      alert(`实例 "${resp.bot_id ? `Bot #${resp.bot_id}` : resp.instance_id}" 已创建，你可以在对话中添加它`);
    } catch (err) {
      alert(`创建失败: ${err instanceof Error ? err.message : "未知错误"}`);
    } finally {
      setQuickCreating(false);
    }
  };

  const loadCommunityBots = async (keyword?: string) => {
    setCommunityBotsLoading(true);
    try {
      const resp = await api.bot.listCommunityBots(keyword || undefined);
      setCommunityHostedBots(resp.hosted_bots || []);
      setCommunityTemplates(resp.templates || []);
    } catch {
      setCommunityHostedBots([]);
      setCommunityTemplates([]);
    } finally {
      setCommunityBotsLoading(false);
    }
  };

  createResource(() => activeTab() === "community", () => { if (activeTab() === "community") loadCommunityBots(); });

  const handleInstallBot = (botId: string) => {
    setInstallBotId(botId);
    setSelectedConvIds(() => new Set<string>());
    setShowInstallDialog(true);
    loadConvList();
  };

  const loadConvList = async () => {
    setConvListLoading(true);
    try {
      const resp = await api.conversation.list();
      setConvList(resp.list || []);
    } catch {
      setConvList([]);
    } finally {
      setConvListLoading(false);
    }
  };

  const toggleConvSelection = (convId: string) => {
    setSelectedConvIds(prev => {
      const next = new Set(prev);
      if (next.has(convId)) next.delete(convId);
      else next.add(convId);
      return next;
    });
  };

  const handleConfirmInstall = async () => {
    const convs = [...selectedConvIds()];
    if (convs.length === 0) { alert("请至少选择一个会话"); return; }
    try {
      const resp = await api.bot.batchInstall(installBotId(), convs);
      if (resp.failed_convs && resp.failed_convs.length > 0) {
        alert(`安装完成：成功 ${resp.success_count} 个，失败 ${resp.failed_convs.length} 个`);
      } else {
        alert(`Bot 安装成功！已安装到 ${resp.success_count} 个会话`);
      }
      setShowInstallDialog(false);
    } catch {
      alert("Bot 安装失败，请稍后再试");
    }
  };

  // ============ RAG (Knowledge Base) State ============
  const [kbs, setKbs] = createSignal<KBItem[]>([]);
  const [kbsLoading, setKbsLoading] = createSignal(false);
  const [showCreateKB, setShowCreateKB] = createSignal(false);
  const [editingKB, setEditingKB] = createSignal<KBItem | null>(null);
  const [kbName, setKbName] = createSignal("");
  const [kbDesc, setKbDesc] = createSignal("");
  const [kbSourceType, setKbSourceType] = createSignal("manual");
  const [kbSourceConfig, setKbSourceConfig] = createSignal("");

  // Document management inside a KB
  const [docsKbId, setDocsKbId] = createSignal<string | null>(null);
  const [docs, setDocs] = createSignal<DocItem[]>([]);
  const [docsLoading, setDocsLoading] = createSignal(false);
  const [uploading, setUploading] = createSignal(false);
  const [syncingKbId, setSyncingKbId] = createSignal<string | null>(null);
  const [syncStatus, setSyncStatus] = createSignal("");

  const loadKBs = async () => {
    setKbsLoading(true);
    try {
      const resp = await api.rag.listKBs();
      setKbs(resp.list || []);
    } catch {
      setKbs([]);
    } finally {
      setKbsLoading(false);
    }
  };

  createResource(() => activeTab() === "rag", () => { if (activeTab() === "rag") loadKBs(); });

  const handleCreateKB = async (e: Event) => {
    e.preventDefault();
    if (!kbName().trim()) return;
    try {
      const data: CreateKBReq = { name: kbName().trim() };
      if (kbDesc().trim()) data.description = kbDesc().trim();
      if (kbSourceType().trim()) data.source_type = kbSourceType().trim();
      if (kbSourceConfig().trim()) data.source_config = kbSourceConfig().trim();
      await api.rag.createKB(data);
      resetKBForm();
      loadKBs();
    } catch (err) {
      alert(`知识库创建失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const handleEditKB = async (e: Event) => {
    e.preventDefault();
    const kb = editingKB();
    if (!kb) return;
    try {
      const data: Partial<CreateKBReq> = {};
      if (kbName().trim()) data.name = kbName().trim();
      if (kbDesc().trim()) data.description = kbDesc().trim();
      if (kbSourceType().trim()) data.source_type = kbSourceType().trim();
      if (kbSourceConfig().trim()) data.source_config = kbSourceConfig().trim();
      await api.rag.updateKB(kb.kb_id, data);
      resetKBForm();
      loadKBs();
    } catch (err) {
      alert(`知识库更新失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const resetKBForm = () => {
    setShowCreateKB(false);
    setEditingKB(null);
    setKbName("");
    setKbDesc("");
    setKbSourceType("manual");
    setKbSourceConfig("");
  };

  const startEditKB = (kb: KBItem) => {
    setEditingKB(kb);
    setKbName(kb.name);
    setKbDesc(kb.description || "");
    setKbSourceType(kb.source_type || "manual");
    setKbSourceConfig("");
    setShowCreateKB(true);
  };

  const handleDeleteKB = async (kbId: string) => {
    if (!confirm("确定要删除这个知识库吗？所有关联文档将被永久删除。")) return;
    try {
      await api.rag.deleteKB(kbId);
      loadKBs();
    } catch (err) {
      alert(`删除失败: ${err instanceof Error ? err.message : "未知错误"}`);
    }
  };

  const loadDocs = async (kbId: string) => {
    setDocsKbId(kbId);
    setDocsLoading(true);
    try {
      const resp = await api.rag.listDocs(kbId);
      setDocs(resp.list || []);
    } catch {
      setDocs([]);
    } finally {
      setDocsLoading(false);
    }
  };

  const handleUploadDoc = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file || !docsKbId()) return;
    setUploading(true);
    try {
      const { doc_id } = await api.rag.uploadDoc(docsKbId()!, file);
      await loadDocs(docsKbId()!);

      // Poll status every 2s until READY or FAILED (max 60s)
      const maxRetries = 30;
      for (let i = 0; i < maxRetries; i++) {
        await new Promise((r) => setTimeout(r, 2000));
        try {
          const s = await api.rag.getDocStatus(docsKbId()!, doc_id);
          if (s.doc.status === "READY" || s.doc.status === "FAILED") break;
        } catch {
          // polling error, continue
        }
      }
      await loadDocs(docsKbId()!);
    } catch (err) {
      alert(`文档上传失败: ${err instanceof Error ? err.message : "未知错误"}`);
    } finally {
      setUploading(false);
      input.value = "";
    }
  };

  const handleDeleteDoc = async (docId: string) => {
    if (!confirm("确定要删除这个文档吗？")) return;
    try {
      await api.rag.deleteDoc(docsKbId()!, docId);
      await loadDocs(docsKbId()!);
    } catch {
      alert("文档删除失败");
    }
  };

  const handleTriggerSync = async (kbId: string) => {
    setSyncingKbId(kbId);
    try {
      const resp = await api.rag.triggerSync(kbId);
      setSyncStatus(`同步已触发: ${resp.sync_id}`);
    } catch (err) {
      setSyncStatus(`同步失败: ${err instanceof Error ? err.message : "未知错误"}`);
    } finally {
      setSyncingKbId(null);
    }
  };

  const resolveUrl = (url?: string) => {
    if (!url) return "";
    if (url.startsWith("http://") || url.startsWith("https://")) return url;
    return getServerApiBase() + url;
  };

  return (
    <div class="h-screen flex flex-col bg-bg">
      {/* Header */}
      <div class="h-14 px-4 border-b border-border flex items-center gap-3 shrink-0 bg-bg-secondary/50">
        <button
          onClick={() => appNav.goChat()}
          class="p-1.5 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text"
        >
          <ArrowLeft size={18} />
        </button>
        <h1 class="text-base font-semibold text-text">设置</h1>
      </div>

      <div class="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div class="w-48 border-r border-border p-3 space-y-1 shrink-0 overflow-y-auto">
          <button
            onClick={() => { setActiveTab("general"); setDocsKbId(null); }}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "general" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><Palette size={16} />通用</button>
          <button
            onClick={() => { setActiveTab("servers"); setDocsKbId(null); }}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "servers" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><Server size={16} />服务器</button>
          <button
            onClick={() => { setActiveTab("account"); setDocsKbId(null); }}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "account" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><User size={16} />账号</button>
          <button
            onClick={() => { setActiveTab("templates"); setDocsKbId(null); }}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "templates" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><Bot size={16} />Bot 模板</button>
          <button
            onClick={() => { setActiveTab("instances"); setDocsKbId(null); }}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "instances" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><Cpu size={16} />Bot 实例</button>
          <button
            onClick={() => { setActiveTab("community"); setDocsKbId(null); }}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "community" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><Globe size={16} />Bot 社区</button>
          <button
            onClick={() => setActiveTab("rag")}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "rag" ? "bg-primary/10 text-primary" : "text-text-secondary hover:bg-surface hover:text-text"
            }`}><Brain size={16} />知识库</button>
        </div>

        {/* Content */}
        <div class="flex-1 overflow-y-auto p-6">
          {/* ======== General Tab ======== */}
          <Show when={activeTab() === "general"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">通用设置</h2>
                <div class="space-y-4">
                  <div class="bg-surface rounded-xl p-4 border border-border">
                    <div class="flex items-center justify-between">
                      <div class="flex items-center gap-3">
                        <Palette size={18} class="text-text-muted" />
                        <div>
                          <p class="text-sm font-medium text-text">浅色模式</p>
                          <p class="text-xs text-text-muted">切换亮色/暗色主题</p>
                        </div>
                      </div>
                      <button
                        onClick={themeStore.toggleTheme}
                        class={`w-10 h-6 rounded-full relative cursor-pointer transition-colors ${
                          themeStore.theme() === "light" ? "bg-primary" : "bg-bg-tertiary"
                        }`}
                      >
                        <div
                          class={`w-4 h-4 rounded-full bg-white absolute top-1 transition-all ${
                            themeStore.theme() === "light" ? "left-5" : "left-1"
                          }`}
                        />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Show>

          {/* ======== Servers Tab ======== */}
          <Show when={activeTab() === "servers"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">服务器管理</h2>
                <form onSubmit={handleAddServer} class="flex gap-2 mb-4">
                  <input
                    type="text"
                    value={newServerUrl()}
                    onInput={(e) => setNewServerUrl(e.currentTarget.value)}
                    placeholder="输入服务器地址，如 http://192.168.1.10:8090"
                    class="flex-1 px-3 py-2 bg-surface border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                  />
                  <button
                    type="submit"
                    disabled={addingServer()}
                    class="px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors disabled:opacity-50 shrink-0 flex items-center gap-1.5"
                  >
                    <Show when={addingServer()}><span class="animate-spin">&#9696;</span></Show>
                    <Plus size={14} />添加
                  </button>
                </form>
                <Show when={serverError()}>
                  <p class="text-xs text-danger mb-3">{serverError()}</p>
                </Show>
                <div class="space-y-2">
                  <For each={servers()}>
                    {(server) => (
                      <div
                        class={`bg-surface rounded-xl p-4 border cursor-pointer transition-colors ${
                          activeServer()?.id === server.id ? "border-primary bg-primary/5" : "border-border hover:border-border-hover"
                        }`}
                        onClick={() => handleSelectServer(server)}
                      >
                        <div class="flex items-center justify-between">
                          <div class="flex items-center gap-3">
                            <div class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                              <Server size={14} class="text-primary" />
                            </div>
                            <div>
                              <p class="text-sm font-medium text-text">{server.name}</p>
                              <p class="text-xs text-text-muted">{server.apiUrl}</p>
                            </div>
                          </div>
                          <div class="flex items-center gap-2">
                            <Show when={activeServer()?.id === server.id}>
                              <Check size={14} class="text-primary" />
                            </Show>
                            <button
                              onClick={(e) => { e.stopPropagation(); handleRemoveServer(server.id); }}
                              class="p-1 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger"
                            >
                              <Trash2 size={14} />
                            </button>
                          </div>
                        </div>
                      </div>
                    )}
                  </For>
                </div>
              </div>
            </div>
          </Show>

          {/* ======== Account Tab ======== */}
          <Show when={activeTab() === "account"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">账号管理</h2>
                <div class="bg-surface rounded-xl p-4 border border-border">
                  <div class="flex items-center gap-4 mb-4 pb-4 border-b border-border">
                    <label class="relative cursor-pointer">
                      <div class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden">
                        <Show when={authStore.avatar()} fallback={<User size={28} class="text-primary" />}>
                          <img src={resolveUrl(authStore.avatar())} alt="avatar" class="w-full h-full object-cover" />
                        </Show>
                      </div>
                      <input type="file" accept="image/*" class="hidden" onChange={handleAvatarUpload} />
                    </label>
                    <div class="flex-1">
                      <div class="flex items-center gap-2">
                        <Show when={editingName()} fallback={
                          <p class="text-base font-semibold text-text">{authStore.name() || "未设置"}</p>
                        }>
                          <input
                            type="text"
                            value={newName()}
                            onInput={(e) => setNewName(e.currentTarget.value)}
                            class="flex-1 px-2 py-1 bg-bg border border-border rounded-lg text-sm focus:outline-none focus:border-primary"
                            onKeyDown={(e) => { if (e.key === "Enter") handleSaveName(); }}
                          />
                        </Show>
                        <Show when={editingName()} fallback={
                          <button onClick={() => { setNewName(authStore.name() || ""); setEditingName(true); }}
                            class="p-1 hover:bg-primary/10 rounded transition-colors text-text-muted"><Edit size={14} /></button>
                        }>
                          <button onClick={handleSaveName} class="p-1 hover:bg-primary/10 rounded transition-colors text-primary"><Check size={14} /></button>
                          <button onClick={() => setEditingName(false)} class="p-1 hover:bg-danger/10 rounded transition-colors text-text-muted"><X size={14} /></button>
                        </Show>
                      </div>
                      <p class="text-xs text-text-muted mt-1">UID: {authStore.uid() || "-"}</p>
                    </div>
                  </div>
                  <button
                    onClick={handleLogout}
                    class="w-full px-4 py-2.5 bg-danger hover:bg-danger/80 text-white rounded-xl text-sm font-medium transition-colors flex items-center justify-center gap-2"
                  >
                    <LogOut size={14} />退出登录
                  </button>
                </div>
              </div>
            </div>
          </Show>

          {/* ======== Templates Tab ======== */}
          <Show when={activeTab() === "templates"}>
            <div class="max-w-lg space-y-6">
              <div class="flex items-center justify-between">
                <h2 class="text-lg font-semibold text-text">Bot 模板</h2>
                <button
                  onClick={() => { resetTemplateForm(); setShowCreateTemplate(true); }}
                  class="flex items-center gap-1.5 px-3 py-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors"
                ><Plus size={14} />创建模板</button>
              </div>
              <Show when={templatesLoading()}><div class="text-center py-8 text-text-muted text-sm">加载中...</div></Show>
              <Show when={!templatesLoading() && templates().length === 0}>
                <div class="text-center py-12 text-text-muted text-sm">
                  <Bot size={40} class="mx-auto mb-3 text-text-muted/30" /><p>还没有创建任何模板</p>
                </div>
              </Show>
              <div class="space-y-2">
                <For each={templates()}>
                  {(tpl) => (
                    <div class="bg-surface rounded-xl p-4 border border-border">
                      <div class="flex items-center gap-3 mb-2">
                        <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                          <Bot size={18} class="text-primary" />
                        </div>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-semibold text-text truncate">{tpl.name}
                            <Show when={tpl.is_official}><span class="ml-1 text-[10px] bg-accent/20 text-accent px-1.5 py-0.5 rounded-full">官方</span></Show>
                          </p>
                          <p class="text-xs text-text-muted truncate">{tpl.description || "暂无描述"}</p>
                        </div>
                        <div class="flex items-center gap-1">
                          <button onClick={() => startEditTemplate(tpl)} class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text" title="编辑"><Edit size={14} /></button>
                          <button onClick={() => handleDeleteTemplate(tpl.template_id)} class="p-1.5 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger" title="删除"><Trash2 size={14} /></button>
                        </div>
                      </div>
                      <div class="bg-bg rounded-lg p-2.5 space-y-1 text-xs font-mono">
                        <div class="flex items-center justify-between"><span class="text-text-muted">ID</span><span class="text-text">{tpl.template_id}</span></div>
                        <Show when={tpl.category}><div class="flex items-center justify-between"><span class="text-text-muted">分类</span><span class="text-text">{tpl.category}</span></div></Show>
                        <div class="flex items-center justify-between"><span class="text-text-muted">版本</span><span class="text-text">{tpl.version || "-"}</span></div>
                        <div class="flex items-center justify-between"><span class="text-text-muted">状态</span><span class="text-text">{tpl.status || "active"}</span></div>
                      </div>
                    </div>
                  )}
                </For>
              </div>
            </div>
          </Show>

          {/* ======== Instances Tab ======== */}
          <Show when={activeTab() === "instances"}>
            <div class="max-w-lg space-y-6">
              <div class="flex items-center justify-between">
                <h2 class="text-lg font-semibold text-text">Bot 实例</h2>
                <button
                  onClick={() => { resetInstanceForm(); setShowCreateInstance(true); }}
                  class="flex items-center gap-1.5 px-3 py-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors"
                ><Plus size={14} />创建实例</button>
              </div>
              <Show when={instancesLoading()}><div class="text-center py-8 text-text-muted text-sm">加载中...</div></Show>
              <Show when={!instancesLoading() && instances().length === 0}>
                <div class="text-center py-12 text-text-muted text-sm">
                  <Cpu size={40} class="mx-auto mb-3 text-text-muted/30" /><p>还没有创建任何实例</p>
                </div>
              </Show>
              <div class="space-y-2">
                <For each={instances()}>
                  {(inst) => (
                    <div class="bg-surface rounded-xl p-4 border border-border">
                      <div class="flex items-center gap-3 mb-2">
                        <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                          <Cpu size={18} class="text-primary" />
                        </div>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-semibold text-text truncate">{inst.name}</p>
                          <p class="text-xs text-text-muted truncate">模板 #{inst.template_id} | Bot #{inst.bot_id}</p>
                        </div>
                        <div class="flex items-center gap-1">
                          <button onClick={() => startEditInstance(inst)} class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text" title="编辑"><Edit size={14} /></button>
                          <button onClick={() => handleDeleteInstance(inst.instance_id)} class="p-1.5 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger" title="删除"><Trash2 size={14} /></button>
                        </div>
                      </div>
                      <div class="bg-bg rounded-lg p-2.5 space-y-1 text-xs font-mono">
                        <div class="flex items-center justify-between"><span class="text-text-muted">实例 ID</span><span class="text-text">{inst.instance_id}</span></div>
                        <div class="flex items-center justify-between"><span class="text-text-muted">Bot ID</span><span class="text-text">{inst.bot_id}</span></div>
                        <div class="flex items-center justify-between"><span class="text-text-muted">模板 ID</span><span class="text-text">{inst.template_id}</span></div>
                        <Show when={inst.model_provider}><div class="flex items-center justify-between"><span class="text-text-muted">模型</span><span class="text-text">{inst.model_provider}/{inst.model_name}</span></div></Show>
                        <div class="flex items-center justify-between"><span class="text-text-muted">托管</span><span class="text-text">{inst.is_self_hosted ? "自托管" : "托管"}</span></div>
                        <div class="flex items-center justify-between"><span class="text-text-muted">状态</span><span class={`text-text ${inst.status === "online" ? "text-green-500" : ""}`}>{inst.status || "offline"}</span></div>
                      </div>
                    </div>
                  )}
                </For>
              </div>
            </div>
          </Show>

          {/* ======== Community Tab ======== */}
          <Show when={activeTab() === "community"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">Bot 社区</h2>
                <p class="text-sm text-text-muted mb-4">发现和安装社区中优秀的 Bot</p>
                <div class="flex gap-2 mb-4">
                  <input type="text" value={communitySearch()} onInput={(e) => setCommunitySearch(e.currentTarget.value)} placeholder="搜索 Bot..."
                    class="flex-1 px-3 py-2 bg-surface border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                    onKeyDown={(e) => { if (e.key === "Enter") loadCommunityBots(communitySearch()); }} />
                  <button onClick={() => loadCommunityBots(communitySearch())}
                    class="px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors shrink-0">搜索</button>
                </div>
              </div>
              <Show when={communityBotsLoading()}><div class="text-center py-8 text-text-muted text-sm">加载中...</div></Show>
              <Show when={!communityBotsLoading() && communityHostedBots().length === 0 && communityTemplates().length === 0}>
                <div class="text-center py-12 text-text-muted text-sm"><Globe size={40} class="mx-auto mb-3 text-text-muted/30" /><p>没有找到 Bot</p></div>
              </Show>

              {/* Hosted Bots */}
              <Show when={communityHostedBots().length > 0}>
                <div class="space-y-2">
                  <h3 class="text-sm font-semibold text-text-muted">托管实例</h3>
                  <For each={communityHostedBots()}>
                    {(bot) => (
                      <div class="bg-surface rounded-xl p-4 border border-border">
                        <div class="flex items-center gap-3 mb-2">
                          <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0"><Bot size={18} class="text-primary" /></div>
                          <div class="flex-1 min-w-0"><p class="text-sm font-semibold text-text truncate">{bot.name}</p><p class="text-xs text-text-muted truncate">{bot.description || "暂无描述"}</p></div>
                          <button onClick={() => handleInstallBot(bot.bot_id)} class="px-3 py-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors shrink-0 flex items-center gap-1"><Plus size={12} />安装</button>
                        </div>
                        <div class="bg-bg rounded-lg p-2.5 grid grid-cols-2 gap-1 text-xs text-text-muted">
                          <div>Bot ID: {bot.bot_id}</div>
                          <div>实例: {bot.installed_count || 0}</div>
                        </div>
                      </div>
                    )}
                  </For>
                </div>
              </Show>

              {/* Templates */}
              <Show when={communityTemplates().length > 0}>
                <div class="space-y-2">
                  <h3 class="text-sm font-semibold text-text-muted">Bot 模板</h3>
                  <For each={communityTemplates()}>
                    {(tpl) => (
                      <div class="bg-surface rounded-xl p-4 border border-border">
                        <div class="flex items-center gap-3 mb-2">
                          <div class="w-10 h-10 rounded-xl bg-accent/10 flex items-center justify-center shrink-0"><Globe size={18} class="text-accent" /></div>
                          <div class="flex-1 min-w-0">
                            <div class="flex items-center gap-2"><p class="text-sm font-semibold text-text truncate">{tpl.name}</p><Show when={tpl.is_official}><span class="text-[10px] px-1.5 py-0.5 rounded-full bg-primary/10 text-primary shrink-0">官方</span></Show></div>
                            <p class="text-xs text-text-muted truncate">{tpl.description || "暂无描述"}</p>
                          </div>
                          <button onClick={() => openQuickCreate(tpl)} class="px-3 py-1.5 bg-accent hover:bg-accent/80 text-white rounded-lg text-xs font-medium transition-colors shrink-0">创建实例</button>
                        </div>
                        <div class="bg-bg rounded-lg p-2.5 grid grid-cols-2 gap-1 text-xs text-text-muted">
                          <div>模板 ID: {tpl.template_id}</div>
                          <div>分类: {tpl.category || "-"}</div>
                        </div>
                      </div>
                    )}
                  </For>
                </div>
              </Show>
            </div>
          </Show>

          {/* ======== Community Quick-Create Modal ======== */}
          <Show when={showQuickCreate()}>
            <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowQuickCreate(false)}>
              <div class="bg-surface rounded-2xl p-6 w-full max-w-sm shadow-xl border border-border mx-4" onClick={e => e.stopPropagation()}>
                <h3 class="text-lg font-semibold text-text mb-4">从模板创建实例</h3>
                <div class="space-y-3">
                  <div>
                    <p class="text-xs text-text-muted mb-1">模板名称</p>
                    <div class="bg-bg rounded-lg px-3 py-2 text-xs text-text-secondary">{quickCreateTemplate()?.name}</div>
                  </div>
                  <div>
                    <label class="text-xs text-text-muted">实例名称</label>
                    <input
                      type="text"
                      value={quickInstName()}
                      onInput={(e) => setQuickInstName(e.currentTarget.value)}
                      placeholder="输入实例名称"
                      class="w-full bg-bg text-text border border-border rounded-lg px-3 py-2 text-sm mt-1 focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>
                  <div>
                    <label class="text-xs text-text-muted">模型提供商（可选）</label>
                    <input
                      type="text"
                      value={quickModelProvider()}
                      onInput={(e) => setQuickModelProvider(e.currentTarget.value)}
                      placeholder="如 openai / qwen"
                      class="w-full bg-bg text-text border border-border rounded-lg px-3 py-2 text-sm mt-1 focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>
                  <div>
                    <label class="text-xs text-text-muted">模型名称（可选）</label>
                    <input
                      type="text"
                      value={quickModelName()}
                      onInput={(e) => setQuickModelName(e.currentTarget.value)}
                      placeholder="如 gpt-4o / qwen-max"
                      class="w-full bg-bg text-text border border-border rounded-lg px-3 py-2 text-sm mt-1 focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>
                </div>
                <div class="flex items-center gap-3 mt-5">
                  <button onClick={() => setShowQuickCreate(false)} disabled={quickCreating()} class="flex-1 px-4 py-2 bg-bg hover:bg-bg/80 text-text-secondary rounded-lg text-sm font-medium transition-colors disabled:opacity-50">取消</button>
                  <button onClick={handleQuickCreate} disabled={quickCreating()} class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
                    {quickCreating() ? "创建中..." : "创建实例"}
                  </button>
                </div>
              </div>
            </div>
          </Show>

          {/* ======== RAG Tab ======== */}
          <Show when={activeTab() === "rag"}>
            <div class="max-w-lg space-y-6">
              <div class="flex items-center justify-between">
                <h2 class="text-lg font-semibold text-text">知识库管理</h2>
                <button
                  onClick={() => { resetKBForm(); setShowCreateKB(true); }}
                  class="flex items-center gap-1.5 px-3 py-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors"
                ><Plus size={14} />创建知识库</button>
              </div>
              <Show when={kbsLoading()}><div class="text-center py-8 text-text-muted text-sm">加载中...</div></Show>
              <Show when={!kbsLoading() && kbs().length === 0}>
                <div class="text-center py-12 text-text-muted text-sm"><Brain size={40} class="mx-auto mb-3 text-text-muted/30" /><p>还没有创建任何知识库</p></div>
              </Show>
              <div class="space-y-2">
                <For each={kbs()}>
                  {(kb) => (
                    <div class="bg-surface rounded-xl border border-border overflow-hidden">
                      <div class="p-4" onClick={() => docsKbId() === kb.kb_id ? setDocsKbId(null) : loadDocs(kb.kb_id)}>
                        <div class="flex items-center gap-3 mb-2">
                          <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0"><Brain size={18} class="text-primary" /></div>
                          <div class="flex-1 min-w-0">
                            <p class="text-sm font-semibold text-text truncate">{kb.name}</p>
                            <p class="text-xs text-text-muted truncate">{kb.description || "暂无描述"}</p>
                          </div>
                          <div class="flex items-center gap-1">
                            <button onClick={(e) => { e.stopPropagation(); startEditKB(kb); }} class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text" title="编辑"><Edit size={14} /></button>
                            <button onClick={(e) => { e.stopPropagation(); handleDeleteKB(kb.kb_id); }} class="p-1.5 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger" title="删除"><Trash2 size={14} /></button>
                          </div>
                        </div>
                        <div class="bg-bg rounded-lg p-2.5 grid grid-cols-2 gap-1 text-xs font-mono">
                          <div><span class="text-text-muted">ID: </span><span class="text-text">{kb.kb_id}</span></div>
                          <div><span class="text-text-muted">文档: </span><span class="text-text">{kb.doc_count || 0}</span></div>
                          <div><span class="text-text-muted">源: </span><span class="text-text">{kb.source_type || "manual"}</span></div>
                          <div><span class="text-text-muted">状态: </span><span class="text-text">{kb.status || "active"}</span></div>
                        </div>
                      </div>
                      {/* Document Sub-panel */}
                      <Show when={docsKbId() === kb.kb_id}>
                        <div class="border-t border-border bg-bg px-4 py-3 space-y-2">
                          <div class="flex items-center justify-between">
                            <span class="text-xs font-medium text-text-muted">文档列表</span>
                            <div class="flex items-center gap-2">
                              <Show when={kb.source_type && kb.source_type !== "manual"}>
                                <button
                                  onClick={() => handleTriggerSync(kb.kb_id)}
                                  disabled={syncingKbId() === kb.kb_id}
                                  class="flex items-center gap-1 px-2 py-1 bg-accent/10 hover:bg-accent/20 text-accent rounded-lg text-xs font-medium transition-colors disabled:opacity-50"
                                >
                                  <RefreshCw size={12} class={syncingKbId() === kb.kb_id ? "animate-spin" : ""} />同步
                                </button>
                              </Show>
                              <label class="flex items-center gap-1 px-2 py-1 bg-primary/10 hover:bg-primary/20 text-primary rounded-lg text-xs font-medium cursor-pointer transition-colors">
                                <Upload size={12} />{uploading() && syncingKbId() === kb.kb_id ? "上传中..." : "上传"}
                                <input type="file" class="hidden" accept=".pdf,.txt,.md,.json,.csv,.docx,.xlsx,.pptx" onChange={handleUploadDoc} />
                              </label>
                            </div>
                          </div>
                          <Show when={syncStatus()}>
                            <p class="text-xs text-accent">{syncStatus()}</p>
                          </Show>
                          <Show when={docsLoading()}><p class="text-xs text-text-muted text-center py-4">加载中...</p></Show>
                          <Show when={!docsLoading() && docs().length === 0}>
                            <p class="text-xs text-text-muted text-center py-4">暂无文档，点击"上传"添加</p>
                          </Show>
                          <div class="space-y-1 max-h-48 overflow-y-auto">
                            <For each={docs()}>
                              {(doc) => (
                                <div class="flex items-center justify-between py-1.5 px-2 bg-surface rounded-lg">
                                  <div class="flex items-center gap-2 min-w-0">
                                    <FileText size={12} class="text-text-muted shrink-0" />
                                    <span class="text-xs text-text truncate">{doc.file_name}</span>
                                    <span class={`text-[10px] px-1 py-0.5 rounded-full shrink-0 ${
                                      doc.status === "completed" ? "bg-green-500/10 text-green-500" :
                                      doc.status === "processing" ? "bg-yellow-500/10 text-yellow-500" :
                                      doc.status === "failed" ? "bg-danger/10 text-danger" :
                                      "bg-bg-tertiary/50 text-text-muted"
                                    }`}>{doc.status}</span>
                                    <span class="text-[10px] text-text-muted shrink-0">{doc.chunk_count || 0} 块</span>
                                  </div>
                                  <button onClick={() => handleDeleteDoc(doc.doc_id)} class="p-1 hover:bg-danger/10 rounded transition-colors text-text-muted hover:text-danger shrink-0"><X size={12} /></button>
                                </div>
                              )}
                            </For>
                          </div>
                        </div>
                      </Show>
                    </div>
                  )}
                </For>
              </div>
            </div>
          </Show>
        </div>
      </div>

      {/* ======== Modals ======== */}

      {/* Create/Edit Template Modal */}
      <Show when={showCreateTemplate()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => resetTemplateForm()}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">{editingTemplate() ? "编辑模板" : "创建模板"}</h2>
            <form onSubmit={editingTemplate() ? handleEditTemplate : handleCreateTemplate} class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">名称 *</label>
                <input type="text" value={tplName()} onInput={(e) => setTplName(e.currentTarget.value)} placeholder="模板名称" required
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">描述</label>
                <input type="text" value={tplDesc()} onInput={(e) => setTplDesc(e.currentTarget.value)} placeholder="简要描述模板功能"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">分类</label>
                <input type="text" value={tplCategory()} onInput={(e) => setTplCategory(e.currentTarget.value)} placeholder="如 chat, assistant, tool"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">System Prompt</label>
                <textarea value={tplSystemPrompt()} onInput={(e) => setTplSystemPrompt(e.currentTarget.value)} placeholder="定义 Bot 的系统提示词..." rows={3}
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors resize-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">欢迎语</label>
                <input type="text" value={tplWelcomeMsg()} onInput={(e) => setTplWelcomeMsg(e.currentTarget.value)} placeholder="Bot 首次对话时的欢迎语"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">标签 (逗号分隔)</label>
                <input type="text" value={tplTags()} onInput={(e) => setTplTags(e.currentTarget.value)} placeholder="如: ai, helpful, coding"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">支持的模型</label>
                <input type="text" value={tplSupportedModels()} onInput={(e) => setTplSupportedModels(e.currentTarget.value)} placeholder="逗号分隔的模型列表"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">工具定义 (JSON)</label>
                <textarea value={tplTools()} onInput={(e) => setTplTools(e.currentTarget.value)} placeholder='JSON tools 定义...' rows={3}
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm font-mono text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors resize-none" />
              </div>
              <div class="flex gap-2 pt-1">
                <button type="button" onClick={() => resetTemplateForm()} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">{editingTemplate() ? "保存" : "创建"}</button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Create/Edit Instance Modal */}
      <Show when={showCreateInstance()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => resetInstanceForm()}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">{editingInstance() ? "编辑实例" : "创建实例"}</h2>
            <form onSubmit={editingInstance() ? handleEditInstance : handleCreateInstance} class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">模板 ID *</label>
                <input type="number" value={instTemplateId()} onInput={(e) => setInstTemplateId(e.currentTarget.value)} placeholder="模板 ID" required disabled={!!editingInstance()}
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors disabled:opacity-50" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">实例名称</label>
                <input type="text" value={instName()} onInput={(e) => setInstName(e.currentTarget.value)} placeholder="可选，默认使用模板名称"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">模型提供商</label>
                <input type="text" value={instModelProvider()} onInput={(e) => setInstModelProvider(e.currentTarget.value)} placeholder="如 openai, anthropic, deepseek"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">模型名称</label>
                <input type="text" value={instModelName()} onInput={(e) => setInstModelName(e.currentTarget.value)} placeholder="如 gpt-4, claude-3-opus"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">API Key</label>
                <input type="password" value={instApiKey()} onInput={(e) => setInstApiKey(e.currentTarget.value)} placeholder="模型提供商的 API Key"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">API Base URL</label>
                <input type="text" value={instApiBaseUrl()} onInput={(e) => setInstApiBaseUrl(e.currentTarget.value)} placeholder="可选，自定义 API 端点"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div class="flex items-center gap-2">
                <input type="checkbox" id="selfHosted" checked={instSelfHosted()} onChange={(e) => setInstSelfHosted(e.currentTarget.checked)}
                  class="rounded border-border" />
                <label for="selfHosted" class="text-xs text-text-muted">自托管模式</label>
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">知识库配置 (JSON)</label>
                <textarea value={instKbConfig()} onInput={(e) => setInstKbConfig(e.currentTarget.value)} placeholder='{"kb_ids": ["kb_xxx"]}' rows={2}
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm font-mono text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors resize-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">实例配置 (JSON)</label>
                <textarea value={instConfig()} onInput={(e) => setInstConfig(e.currentTarget.value)} placeholder='{"temperature": 0.7}' rows={2}
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm font-mono text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors resize-none" />
              </div>
              <div class="flex gap-2 pt-1">
                <button type="button" onClick={() => resetInstanceForm()} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">{editingInstance() ? "保存" : "创建"}</button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Create/Edit KB Modal */}
      <Show when={showCreateKB()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => resetKBForm()}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">{editingKB() ? "编辑知识库" : "创建知识库"}</h2>
            <form onSubmit={editingKB() ? handleEditKB : handleCreateKB} class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">名称 *</label>
                <input type="text" value={kbName()} onInput={(e) => setKbName(e.currentTarget.value)} placeholder="知识库名称" required
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">描述</label>
                <input type="text" value={kbDesc()} onInput={(e) => setKbDesc(e.currentTarget.value)} placeholder="知识库用途描述"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors" />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">数据源类型</label>
                <select value={kbSourceType()} onChange={(e) => setKbSourceType(e.currentTarget.value)}
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text focus:outline-none focus:border-primary transition-colors">
                  <option value="manual">手动上传</option>
                  <option value="feishu">飞书文档</option>
                  <option value="github">GitHub</option>
                  <option value="notion">Notion</option>
                  <option value="web">网页抓取</option>
                </select>
              </div>
              <Show when={kbSourceType() !== "manual"}>
                <div>
                  <label class="block text-xs font-medium text-text-muted mb-1">数据源配置 (JSON)</label>
                  <textarea value={kbSourceConfig()} onInput={(e) => setKbSourceConfig(e.currentTarget.value)} placeholder='{"space_id": "xxx"}' rows={2}
                    class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm font-mono text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors resize-none" />
                </div>
              </Show>
              <div class="flex gap-2 pt-1">
                <button type="button" onClick={() => resetKBForm()} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">{editingKB() ? "保存" : "创建"}</button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Install Bot Dialog */}
      <Show when={showInstallDialog()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowInstallDialog(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-md mx-4 max-h-[80vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-2">选择安装会话</h2>
            <p class="text-xs text-text-muted mb-3">请选择要安装 Bot 的会话（可多选）</p>
            <Show when={convListLoading()}><p class="text-sm text-text-muted text-center py-8">加载中...</p></Show>
            <Show when={!convListLoading()}>
              <div class="flex-1 overflow-y-auto space-y-1 mb-4">
                <For each={convList()}>
                  {(conv: { conv_id: string; name: string; type: string; avatar: string }) => (
                    <div
                      class={`flex items-center gap-3 p-3 rounded-xl cursor-pointer transition-colors ${selectedConvIds().has(conv.conv_id) ? "bg-primary/10 border border-primary/30" : "hover:bg-bg border border-transparent"}`}
                      onClick={() => toggleConvSelection(conv.conv_id)}
                    >
                      <div class={`w-5 h-5 rounded-md border-2 flex items-center justify-center shrink-0 transition-colors ${selectedConvIds().has(conv.conv_id) ? "bg-primary border-primary" : "border-border"}`}>
                        <Show when={selectedConvIds().has(conv.conv_id)}><Check size={12} class="text-white" /></Show>
                      </div>
                      <div class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"><span class="text-xs text-primary font-bold">{conv.name.charAt(0)}</span></div>
                      <div class="flex-1 min-w-0"><p class="text-sm font-medium text-text truncate">{conv.name}</p><p class="text-xs text-text-muted">{conv.type === "group" ? "群聊" : "单聊"}</p></div>
                    </div>
                  )}
                </For>
                <Show when={convList().length === 0}><p class="text-sm text-text-muted text-center py-8">暂无会话</p></Show>
              </div>
            </Show>
            <div class="flex gap-2 pt-2 border-t border-border">
              <button onClick={() => setShowInstallDialog(false)} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
              <button onClick={handleConfirmInstall} class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors" disabled={convListLoading()}>安装到 {selectedConvIds().size} 个会话</button>
            </div>
          </div>
        </div>
      </Show>
    </div>
  );
}