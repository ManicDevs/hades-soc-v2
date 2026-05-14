// ============================================================================
// Hades-V2 Comprehensive Type Definitions
// ============================================================================

// ============================================================================
// User & Authentication Types
// ============================================================================

export type UserRole = 'Administrator' | 'Security Analyst' | 'User' | 'Analyst' | 'ReadOnly';

export type Permission =
  | 'read'
  | 'write'
  | 'admin'
  | 'delete'
  | 'execute'
  | 'manage_users'
  | 'view_logs'
  | 'manage_threats'
  | 'manage_policies'
  | 'view_reports'
  | 'configure_settings';

export interface User {
  id: number | string;
  username: string;
  email: string;
  role: UserRole;
  status: 'active' | 'inactive' | 'suspended' | 'pending';
  created_at?: string;
  last_login?: string;
  permissions?: Permission[];
  avatar?: string;
  department?: string;
  mfa_enabled?: boolean;
}

export interface UserProfile extends User {
  first_name?: string;
  last_name?: string;
  phone?: string;
  organization?: string;
  timezone?: string;
  preferences?: UserPreferences;
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  notifications: NotificationPreferences;
  dashboard_layout?: Record<string, unknown>;
}

export interface NotificationPreferences {
  email: boolean;
  push: boolean;
  slack: boolean;
  threat_alerts: boolean;
  system_alerts: boolean;
  weekly_reports: boolean;
}

export interface AuthTokens {
  access_token: string;
  refresh_token?: string;
  expires_at: number;
  token_type: 'Bearer' | 'Basic';
}

export interface LoginCredentials {
  username: string;
  password: string;
  remember_me?: boolean;
  mfa_code?: string;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

// ============================================================================
// API Response Types
// ============================================================================

export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
  timestamp?: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
  status_code: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: Pagination;
}

export interface Pagination {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ApiListResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ApiFilterParams {
  page?: number;
  limit?: number;
  sort?: string;
  order?: 'asc' | 'desc';
  search?: string;
  filter?: Record<string, unknown>;
}

export interface ApiEndpointConfig {
  baseURL: string;
  timeout: number;
  retryAttempts: number;
  retryDelay: number;
}

// ============================================================================
// Dashboard Types
// ============================================================================

export interface DashboardMetrics {
  security_score: number;
  active_threats: number;
  blocked_attacks: number;
  system_health: number;
  active_users: number;
  cpu_usage?: number;
  memory_usage?: number;
  disk_usage?: number;
  network_throughput?: number;
}

export interface DashboardActivity {
  id: number;
  type: 'threat' | 'user' | 'system' | 'security' | 'policy';
  message: string;
  time: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  timestamp?: string;
  metadata?: Record<string, unknown>;
}

export interface SystemStatus {
  api: 'online' | 'offline' | 'degraded';
  database: 'connected' | 'disconnected' | 'degraded';
  workers: string;
  memory: string;
  cpu: string;
  uptime?: number;
  version?: string;
  services?: Record<string, ServiceStatus>;
}

export interface ServiceStatus {
  name: string;
  status: 'healthy' | 'unhealthy' | 'degraded';
  last_check: string;
  response_time?: number;
}

// ============================================================================
// Threat Types
// ============================================================================

export type ThreatSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';
export type ThreatStatus = 'open' | 'investigating' | 'mitigated' | 'blocked' | 'resolved' | 'monitoring';
export type ThreatType = 'malware' | 'phishing' | 'ddos' | 'sql_injection' | 'xss' | 'ransomware' | 'apt' | 'insider' | 'other';

export interface Threat {
  id: number | string;
  type: ThreatType;
  severity: ThreatSeverity;
  title: string;
  description: string;
  source: ThreatSource;
  status: ThreatStatus;
  timestamp: string;
  last_updated?: string;
  impact?: ThreatImpact;
  indicators?: ThreatIndicator[];
  mitre_ttp?: string[];
  tags?: string[];
  assigned_to?: string;
  notes?: string;
}

export interface ThreatSource {
  ip_address: string;
  country?: string;
  asn?: string;
  domain?: string;
  url?: string;
  port?: number;
  protocol?: string;
}

export interface ThreatImpact {
  risk_score: number;
  affected_assets: string[];
  business_impact: string;
  data_classification: string;
  financial_impact?: number;
  reputational_impact?: string;
}

export interface ThreatIndicator {
  type: 'ip' | 'domain' | 'hash' | 'url' | 'file' | 'behavior';
  value: string;
  confidence: number;
  source: string;
  first_seen?: string;
  last_seen?: string;
}

export interface ThreatStats {
  total_threats: number;
  by_severity: Record<ThreatSeverity, number>;
  by_status: Record<ThreatStatus, number>;
  by_type: Record<ThreatType, number>;
  trend?: ThreatTrend[];
}

export interface ThreatTrend {
  date: string;
  count: number;
  severity: ThreatSeverity;
}

export interface ThreatFilter {
  severity?: ThreatSeverity[];
  status?: ThreatStatus[];
  type?: ThreatType[];
  date_from?: string;
  date_to?: string;
  source?: string;
  assigned_to?: string;
}

// ============================================================================
// Security Types
// ============================================================================

export type SecurityPolicyType =
  | 'access_control'
  | 'data_protection'
  | 'network_security'
  | 'endpoint_security'
  | 'application_security'
  | 'compliance';

export interface SecurityPolicy {
  id: string;
  name: string;
  type: SecurityPolicyType;
  status: 'active' | 'inactive' | 'draft' | 'pending_review';
  severity: ThreatSeverity;
  description: string;
  rules?: PolicyRule[];
  created_at: string;
  updated_at?: string;
  created_by?: string;
  last_modified_by?: string;
  version?: number;
  compliance_frameworks?: string[];
}

export interface PolicyRule {
  id: string;
  name: string;
  condition: string;
  action: 'allow' | 'deny' | 'log' | 'alert';
  enabled: boolean;
  priority: number;
}

export type VulnerabilityStatus = 'open' | 'in_progress' | 'mitigated' | 'resolved' | 'accepted' | 'false_positive';

export interface Vulnerability {
  id: string;
  name: string;
  description: string;
  severity: ThreatSeverity;
  cvss_score: number;
  cvss_vector?: string;
  status: VulnerabilityStatus;
  discovered_at: string;
  resolved_at?: string;
  affected_systems: string[];
  affected_components?: string[];
  cve_ids?: string[];
  cwe_id?: string;
  references?: string[];
  assigned_to?: string;
  remediation?: string;
  exploit_available?: boolean;
  false_positive?: boolean;
}

export interface SecurityScore {
  overall_score: number;
  last_updated: string;
  components: {
    access_control: number;
    data_protection: number;
    network_security: number;
    vulnerability_management: number;
    endpoint_security?: number;
    application_security?: number;
  };
  trend: 'improving' | 'stable' | 'declining';
  recommendations?: string[];
  historical_scores?: HistoricalScore[];
}

export interface HistoricalScore {
  date: string;
  score: number;
}

// ============================================================================
// Kubernetes Types
// ============================================================================

export type K8sResourceStatus = 'running' | 'pending' | 'failed' | 'succeeded' | 'unknown';

export interface K8sCluster {
  id: string;
  name: string;
  status: K8sResourceStatus;
  nodes: number;
  version: string;
  region: string;
  created_at: string;
  resources: K8sResources;
  cloud_provider?: string;
  master_url?: string;
}

export interface K8sResources {
  cpu: string;
  memory: string;
  storage: string;
  cpu_used?: string;
  memory_used?: string;
}

export interface K8sDeployment {
  id: string;
  name: string;
  namespace: string;
  replicas: number;
  ready: number;
  status: K8sResourceStatus;
  image: string;
  created_at: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  strategy?: string;
}

export interface K8sService {
  id: string;
  name: string;
  namespace: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';
  cluster_ip: string;
  external_ip?: string;
  ports: string[];
  selector?: Record<string, string>;
}

export interface K8sAutoscaler {
  id: string;
  name: string;
  namespace: string;
  min_replicas: number;
  max_replicas: number;
  current_replicas: number;
  target_cpu: number;
  target_memory?: number;
  status: 'active' | 'inactive';
}

export interface K8sPod {
  id: string;
  name: string;
  namespace: string;
  status: K8sResourceStatus;
  node: string;
  ip: string;
  created_at: string;
  containers: K8sContainer[];
  labels?: Record<string, string>;
}

export interface K8sContainer {
  name: string;
  image: string;
  ready: boolean;
  restart_count: number;
  state: string;
  resources?: ContainerResources;
}

export interface ContainerResources {
  limits?: { cpu: string; memory: string };
  requests?: { cpu: string; memory: string };
}

// ============================================================================
// Event Types
// ============================================================================

export type AgentEventType = 'threat' | 'system' | 'network' | 'user' | 'policy' | 'compliance';
export type EventSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface AgentEvent {
  id: string;
  type: AgentEventType;
  timestamp: string;
  severity: EventSeverity;
  source: string;
  agent_id?: string;
  message: string;
  data?: Record<string, unknown>;
  metadata?: EventMetadata;
  tags?: string[];
}

export interface EventMetadata {
  hostname?: string;
  process_name?: string;
  pid?: number;
  user_id?: string;
  correlation_id?: string;
}

export type SIEMEventType =
  | 'intrusion_attempt'
  | 'malware_detection'
  | 'failed_login'
  | 'successful_login'
  | 'file_access'
  | 'network_connection'
  | 'privilege_escalation'
  | 'data_exfiltration'
  | 'lateral_movement'
  | 'command_execution';

export interface SIEMEvent {
  id: string;
  timestamp: string;
  severity: EventSeverity;
  source: string;
  event_type: SIEMEventType;
  details: string;
  status: 'new' | 'investigating' | 'contained' | 'resolved' | 'false_positive';
  raw_log?: string;
  enriched_data?: Record<string, unknown>;
  correlation_id?: string;
}

export interface SIEMAlert {
  id: string;
  title: string;
  severity: ThreatSeverity;
  created_at: string;
  source: string;
  status: 'active' | 'investigating' | 'resolved' | 'dismissed';
  description: string;
  events_count?: number;
  triggered_by?: string;
}

export interface SIEMCorrelation {
  id: string;
  name: string;
  confidence: number;
  events_count: number;
  time_window: number;
  description: string;
  status: 'active' | 'paused';
  rules?: string[];
}

// ============================================================================
// Threat Intelligence Types
// ============================================================================

export interface ThreatIntelligence {
  id: string;
  name: string;
  type: 'apt' | 'malware' | 'campaign' | 'tool' | 'infrastructure';
  confidence: number;
  severity: ThreatSeverity;
  description: string;
  indicators: string[];
  mitre_techniques?: string[];
  associated_actors?: string[];
  first_seen?: string;
  last_updated?: string;
  tags?: string[];
}

export interface ThreatFeed {
  id: string;
  name: string;
  source: string;
  last_updated: string;
  indicators_count: number;
  status: 'active' | 'inactive' | 'error';
  description: string;
  feed_type?: 'malware' | 'threat_actor' | 'vulnerability' | 'network';
  reliability_score?: number;
}

// ============================================================================
// Incident Response Types
// ============================================================================

export type IncidentSeverity = 'critical' | 'high' | 'medium' | 'low';
export type IncidentStatus = 'open' | 'investigating' | 'contained' | 'resolved' | 'closed';
export type PlaybookCategory = 'malware' | 'data_loss' | 'phishing' | 'ddos' | 'insider_threat' | 'ransomware';

export interface Incident {
  id: string;
  title: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  created_at: string;
  updated_at?: string;
  resolved_at?: string;
  affected_systems: string[];
  assigned_to: string;
  description: string;
  timeline?: IncidentTimelineEvent[];
  evidence?: Evidence[];
  response_actions?: string[];
}

export interface IncidentTimelineEvent {
  id: string;
  timestamp: string;
  action: string;
  actor: string;
  details: string;
}

export interface Evidence {
  id: string;
  type: 'log' | 'screenshot' | 'file' | 'network_capture' | 'memory_dump';
  name: string;
  path: string;
  collected_at: string;
  collected_by: string;
  hash?: string;
  size?: number;
}

export interface Playbook {
  id: string;
  name: string;
  category: PlaybookCategory;
  severity: IncidentSeverity;
  status: 'active' | 'draft' | 'archived';
  last_updated: string;
  steps: number;
  estimated_duration: number;
  description: string;
  stages?: PlaybookStage[];
}

export interface PlaybookStage {
  id: string;
  name: string;
  order: number;
  actions: PlaybookAction[];
  estimated_time?: number;
}

export interface PlaybookAction {
  id: string;
  name: string;
  type: 'manual' | 'automated' | 'approval';
  description: string;
  automation_script?: string;
}

export interface ResponseAction {
  id: string;
  name: string;
  category: 'containment' | 'investigation' | 'remediation' | 'communication' | 'recovery';
  automation_level: 'automated' | 'semi_automated' | 'manual';
  execution_time: number;
  description: string;
  prerequisites?: string[];
  risks?: string[];
}

// ============================================================================
// Analytics & ML Types
// ============================================================================

export interface AnalyticsOverview {
  total_events: number;
  critical_alerts: number;
  ml_accuracy: number;
  prediction_confidence: number;
  last_updated: string;
}

export interface MLInsight {
  id: string;
  type: 'anomaly_detection' | 'behavioral_analysis' | 'threat_prediction' | 'pattern_recognition';
  confidence: number;
  description: string;
  severity: EventSeverity;
  timestamp: string;
  affected_assets?: string[];
  recommendation?: string;
}

export interface Prediction {
  id: string;
  type: 'threat_prediction' | 'resource_usage' | 'attack_probability';
  probability: number;
  timeframe: string;
  description: string;
  confidence: number;
  factors?: string[];
}

export interface AnalyticsMetrics {
  detection_rate: number;
  false_positive_rate: number;
  response_time: number;
  coverage: number;
  efficiency: number;
}

export interface TrendData {
  metric: string;
  direction: 'increasing' | 'decreasing' | 'stable';
  change_percentage: number;
  period: string;
  significance: 'high' | 'medium' | 'low';
  historical_data?: { date: string; value: number }[];
}

// ============================================================================
// Threat Hunting Types
// ============================================================================

export type HuntStatus = 'created' | 'active' | 'paused' | 'completed' | 'cancelled';
export type HuntType = 'proactive' | 'reactive' | 'routine' | 'targeted';

export interface ThreatHunt {
  id: string;
  name: string;
  status: HuntStatus;
  type: HuntType;
  started_at: string;
  completed_at?: string;
  progress: number;
  threats_found: number;
  description: string;
  hypothesis?: string;
  creator?: string;
  tags?: string[];
}

export interface HuntIndicator {
  id: string;
  type: 'ioc' | 'ioa' | 'ttp' | 'behavior';
  value: string;
  confidence: number;
  source: string;
  description: string;
  first_seen?: string;
  last_seen?: string;
  tags?: string[];
}

// ============================================================================// Blockchain & Audit Types
// ============================================================================

export interface AuditLog {
  id: string;
  timestamp: string;
  action: string;
  user: string;
  resource: string;
  result: 'success' | 'failure';
  hash: string;
  block_hash: string;
  metadata?: Record<string, unknown>;
  ip_address?: string;
  user_agent?: string;
}

export interface BlockchainBlock {
  id: string;
  hash: string;
  previous_hash: string;
  timestamp: string;
  nonce: number;
  transactions: string[];
  size: string;
  validator?: string;
}

export interface BlockchainTransaction {
  id: string;
  hash: string;
  block_hash: string;
  timestamp: string;
  type: 'audit_log' | 'config_change' | 'user_action' | 'policy_update';
  status: 'pending' | 'confirmed' | 'rejected';
  from?: string;
  to?: string;
  data?: Record<string, unknown>;
}

export interface ChainIntegrity {
  is_valid: boolean;
  last_verified: string;
  total_blocks: number;
  chain_hash: string;
  verification_confidence?: number;
}

// ============================================================================
// Zero Trust Types
// ============================================================================

export type TrustScoreLevel = 'high' | 'medium' | 'low' | 'critical';

export interface ZeroTrustPolicy {
  id: string;
  name: string;
  status: 'active' | 'inactive' | 'draft';
  created_at: string;
  rules_count: number;
  description: string;
  enforcement_level?: string;
  conditions?: PolicyCondition[];
}

export interface PolicyCondition {
  attribute: string;
  operator: 'equals' | 'not_equals' | 'contains' | 'greater_than' | 'less_than';
  value: string | number;
}

export interface AccessRequest {
  id: string;
  user: string;
  resource: string;
  requested_at: string;
  status: 'pending' | 'approved' | 'denied' | 'expired';
  trust_score: number;
  decision?: string;
  decision_reason?: string;
  reviewed_by?: string;
  reviewed_at?: string;
}

export interface TrustScore {
  user_id: string;
  score: number;
  last_updated: string;
  factors: ('authentication' | 'behavior' | 'location' | 'device' | 'time')[];
  history?: TrustScoreHistory[];
}

export interface TrustScoreHistory {
  timestamp: string;
  score: number;
  change_reason: string;
}

export interface NetworkSegment {
  id: string;
  name: string;
  trust_level: TrustScoreLevel;
  devices_count: number;
  isolation_status: 'isolated' | 'semi_isolated' | 'open';
  allowed_sources?: string[];
  allowed_destinations?: string[];
}

// ============================================================================
// Quantum Security Types
// ============================================================================

export type QuantumAlgorithmType = 'key_exchange' | 'encryption' | 'signature' | 'hash';

export interface QuantumAlgorithm {
  id: string;
  name: string;
  type: QuantumAlgorithmType;
  security_level: number;
  key_size: number;
  status: 'active' | 'deprecated' | 'testing';
  description: string;
  performance?: AlgorithmPerformance;
}

export interface AlgorithmPerformance {
  keygen_time: string;
  encap_time: string;
  decap_time: string;
}

export interface QuantumKey {
  id: string;
  algorithm_id: string;
  created_at: string;
  expires_at: string;
  status: 'active' | 'expired' | 'revoked';
  key_size: number;
  public_key?: string;
  encrypted_private_key?: string;
}

export interface QuantumCertificate {
  id: string;
  key_id: string;
  issued_at: string;
  expires_at: string;
  status: 'valid' | 'expired' | 'revoked';
  issuer: string;
  subject?: string;
  serial_number?: string;
}

export interface QuantumMetrics {
  total_keys: number;
  active_keys: number;
  expired_keys: number;
  revoked_keys?: number;
  security_score: number;
  last_rotation: string;
}

// ============================================================================
// Threat Modeling Types
// ============================================================================

export interface ThreatModel {
  id: string;
  name: string;
  version: string;
  coverage: number;
  last_updated: string;
  techniques_count: number;
  description: string;
  framework?: string;
  categories?: string[];
}

export interface ThreatScenario {
  id: string;
  name: string;
  complexity: 'low' | 'medium' | 'high';
  likelihood: 'very_low' | 'low' | 'medium' | 'high' | 'very_high';
  impact: ThreatSeverity;
  duration: number;
  description: string;
  prerequisites?: string[];
  affected_assets?: string[];
  mitigation_steps?: string[];
}

export interface ThreatSimulation {
  id: string;
  name: string;
  type: 'red_team' | 'blue_team' | 'purple_team';
  status: 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
  scheduled_at?: string;
  started_at?: string;
  completed_at?: string;
  duration: number;
  success_rate: number;
  description: string;
  results?: SimulationResult;
}

export interface SimulationResult {
  objectives_achieved: string[];
  vulnerabilities_found: number;
  detection_rate: number;
  response_time: number;
  lessons_learned?: string[];
  recommendations?: string[];
}

export interface Mitigation {
  id: string;
  name: string;
  type: 'technical' | 'administrative' | 'physical';
  status: 'planned' | 'in_progress' | 'implemented';
  effectiveness: number;
  implemented_at?: string;
  description: string;
  cost?: number;
  owner?: string;
}

// ============================================================================
// Form Types
// ============================================================================

export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'number' | 'select' | 'checkbox' | 'radio' | 'textarea' | 'date' | 'datetime';
  required?: boolean;
  placeholder?: string;
  options?: { label: string; value: string | number }[];
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    message?: string;
  };
  defaultValue?: unknown;
}

export interface FormSchema {
  fields: FormField[];
  submitLabel?: string;
  cancelLabel?: string;
  onSubmit?: (data: Record<string, unknown>) => void;
  onCancel?: () => void;
}

// ============================================================================
// Utility Types
// ============================================================================

export type Nullable<T> = T | null;
export type Optional<T> = T | undefined;
export type DeepPartial<T> = { [P in keyof T]?: DeepPartial<T[P]> };

export type AsyncState<T> = {
  data: T | null;
  isLoading: boolean;
  error: string | null;
};

export type SelectOption = {
  label: string;
  value: string | number;
  disabled?: boolean;
};

export type DateRange = {
  start: string;
  end: string;
};

export type ChartDataPoint = {
  name: string;
  value: number;
  color?: string;
};

export type TimeSeriesData = {
  timestamp: string;
  value: number;
};