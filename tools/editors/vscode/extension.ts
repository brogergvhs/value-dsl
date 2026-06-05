import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;

function fileExists(p: string): boolean {
  try {
    return fs.existsSync(p);
  } catch {
    return false;
  }
}

function resolveServerCommand(context: vscode.ExtensionContext): string {
  const config = vscode.workspace.getConfiguration("dsl");
  const configuredPath = config.get<string>("lsp.path");
  const binaryName = process.platform === "win32" ? "dsl-lsp.exe" : "dsl-lsp";

  if (configuredPath && configuredPath.trim().length > 0) {
    return configuredPath.trim();
  }

  const bundledBinary = path.join(context.extensionPath, "bin", binaryName);
  if (fileExists(bundledBinary)) {
    return bundledBinary;
  }

  /**
   * During development, this extension lives at:
   *
   *   value-dsl/tools/editors/vscode
   *
   * So the repo root is three levels up.
   */
  const devRepoBinary = path.resolve(
    context.extensionPath,
    "..",
    "..",
    "..",
    "bin",
    binaryName,
  );

  if (fileExists(devRepoBinary)) {
    return devRepoBinary;
  }

  return binaryName;
}

async function startClient(context: vscode.ExtensionContext) {
  const command = resolveServerCommand(context);

  const outputChannel = vscode.window.createOutputChannel(
    "Value DSL Language Server",
    { log: true },
  );

  const serverOptions: ServerOptions = {
    command,
    args: [],
    options: {
      env: process.env,
    },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: "file", language: "dsl" },
      { scheme: "untitled", language: "dsl" },
    ],
    outputChannel,
  };

  client = new LanguageClient(
    "value-dsl-lsp",
    "Value DSL Language Server",
    serverOptions,
    clientOptions,
  );

  context.subscriptions.push(
    vscode.commands.registerCommand(
      "valueDsl.restartLanguageServer",
      async () => {
        if (client) {
          await client.stop();
          client = undefined;
        }

        await startClient(context);
        vscode.window.showInformationMessage(
          "Value DSL language server restarted.",
        );
      },
    ),
  );

  try {
    outputChannel.appendLine(`Starting Value DSL language server: ${command}`);
    await client.start();
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);

    vscode.window.showErrorMessage(
      `Could not start Value DSL language server: ${message}`,
    );

    outputChannel.appendLine(`Failed to start server: ${message}`);
  }
}

export async function activate(context: vscode.ExtensionContext) {
  await startClient(context);
}

export async function deactivate() {
  if (client) {
    await client.stop();
    client = undefined;
  }
}
