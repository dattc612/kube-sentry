import { ExecException } from "node:child_process";

export interface JobItem {
  ID: number;
  JobName: string;
  CMD: string[];
  Description: string;
}
