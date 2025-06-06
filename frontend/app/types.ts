export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
}

export interface Project {
  id: number;
  name: string;
  description: string;
  status: string;
  user_id: number;
  user?: User;
  created_at: string;
  updated_at: string;
}

export interface FeatureTag {
  tag_name: string;
  feature_id: number;
  created_by_user: number;
}

export interface Feature {
  id: number;
  project_id: number;
  title: string;
  description: string;
  status: "todo" | "in_progress" | "done";
  priority: "low" | "medium" | "high";
  assignee_id: number;
  created_at: string;
  updated_at: string;
  assignee?: User;
  project?: Project;
  tags?: FeatureTag[];
  tags_input?: string; // For form handling
} 