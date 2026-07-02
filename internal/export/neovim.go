package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

func GenerateNeovimLua(exportDir string, colors *theme.Colors, lightMode bool) error {
	var b strings.Builder
	backgroundMode := "dark"
	if lightMode {
		backgroundMode = "light"
	}

	b.WriteString("return {\n")
	b.WriteString("  {\n")
	b.WriteString("    name = \"omarchy-themegen\",\n")
	b.WriteString("    dir = vim.fn.stdpath(\"config\"),\n")
	b.WriteString("    lazy = false,\n")
	b.WriteString("    priority = 1000,\n")
	b.WriteString("    config = function()\n")
	b.WriteString("      local colors = {\n")
	writeLuaColor(&b, "accent", colors.Accent)
	writeLuaColor(&b, "cursor", colors.Cursor)
	writeLuaColor(&b, "foreground", colors.Foreground)
	writeLuaColor(&b, "background", colors.Background)
	writeLuaColor(&b, "selection_foreground", colors.SelectionForeground)
	writeLuaColor(&b, "selection_background", colors.SelectionBackground)
	writeLuaColor(&b, "color0", colors.Color0)
	writeLuaColor(&b, "color1", colors.Color1)
	writeLuaColor(&b, "color2", colors.Color2)
	writeLuaColor(&b, "color3", colors.Color3)
	writeLuaColor(&b, "color4", colors.Color4)
	writeLuaColor(&b, "color5", colors.Color5)
	writeLuaColor(&b, "color6", colors.Color6)
	writeLuaColor(&b, "color7", colors.Color7)
	writeLuaColor(&b, "color8", colors.Color8)
	writeLuaColor(&b, "color9", colors.Color9)
	writeLuaColor(&b, "color10", colors.Color10)
	writeLuaColor(&b, "color11", colors.Color11)
	writeLuaColor(&b, "color12", colors.Color12)
	writeLuaColor(&b, "color13", colors.Color13)
	writeLuaColor(&b, "color14", colors.Color14)
	writeLuaColor(&b, "color15", colors.Color15)
	b.WriteString("      }\n")
	b.WriteString("\n")
	b.WriteString("      local function hl(group, opts)\n")
	b.WriteString("        vim.api.nvim_set_hl(0, group, opts)\n")
	b.WriteString("      end\n")
	b.WriteString("\n")
	b.WriteString("      local function apply()\n")
	b.WriteString("        vim.opt.termguicolors = true\n")
	b.WriteString(fmt.Sprintf("        vim.o.background = %q\n", backgroundMode))
	b.WriteString("        vim.g.colors_name = \"omarchy-themegen\"\n")
	b.WriteString("\n")
	b.WriteString("        hl(\"Normal\", { fg = colors.foreground, bg = colors.background })\n")
	b.WriteString("        hl(\"NormalNC\", { fg = colors.foreground, bg = colors.background })\n")
	b.WriteString("        hl(\"NormalFloat\", { fg = colors.foreground, bg = colors.color0 })\n")
	b.WriteString("        hl(\"FloatBorder\", { fg = colors.accent, bg = colors.color0 })\n")
	b.WriteString("        hl(\"Cursor\", { fg = colors.background, bg = colors.cursor })\n")
	b.WriteString("        hl(\"CursorLine\", { bg = colors.color0 })\n")
	b.WriteString("        hl(\"CursorLineNr\", { fg = colors.accent, bold = true })\n")
	b.WriteString("        hl(\"LineNr\", { fg = colors.color8 })\n")
	b.WriteString("        hl(\"SignColumn\", { fg = colors.color8, bg = colors.background })\n")
	b.WriteString("        hl(\"WinSeparator\", { fg = colors.color8, bg = colors.background })\n")
	b.WriteString("        hl(\"StatusLine\", { fg = colors.foreground, bg = colors.color8 })\n")
	b.WriteString("        hl(\"StatusLineNC\", { fg = colors.color7, bg = colors.color0 })\n")
	b.WriteString("        hl(\"Visual\", { fg = colors.selection_foreground, bg = colors.selection_background })\n")
	b.WriteString("        hl(\"Search\", { fg = colors.background, bg = colors.color3 })\n")
	b.WriteString("        hl(\"IncSearch\", { fg = colors.background, bg = colors.accent })\n")
	b.WriteString("        hl(\"Pmenu\", { fg = colors.foreground, bg = colors.color0 })\n")
	b.WriteString("        hl(\"PmenuSel\", { fg = colors.selection_foreground, bg = colors.selection_background })\n")
	b.WriteString("\n")
	b.WriteString("        hl(\"Comment\", { fg = colors.color8, italic = true })\n")
	b.WriteString("        hl(\"Constant\", { fg = colors.color1 })\n")
	b.WriteString("        hl(\"String\", { fg = colors.color2 })\n")
	b.WriteString("        hl(\"Identifier\", { fg = colors.color5 })\n")
	b.WriteString("        hl(\"Function\", { fg = colors.color4 })\n")
	b.WriteString("        hl(\"Statement\", { fg = colors.color3 })\n")
	b.WriteString("        hl(\"Type\", { fg = colors.color6 })\n")
	b.WriteString("        hl(\"Special\", { fg = colors.accent })\n")
	b.WriteString("        hl(\"Directory\", { fg = colors.color4 })\n")
	b.WriteString("        hl(\"Error\", { fg = colors.color1 })\n")
	b.WriteString("        hl(\"WarningMsg\", { fg = colors.color3 })\n")
	b.WriteString("        hl(\"DiagnosticError\", { fg = colors.color1 })\n")
	b.WriteString("        hl(\"DiagnosticWarn\", { fg = colors.color3 })\n")
	b.WriteString("        hl(\"DiagnosticInfo\", { fg = colors.color4 })\n")
	b.WriteString("        hl(\"DiagnosticHint\", { fg = colors.color6 })\n")
	b.WriteString("      end\n")
	b.WriteString("\n")
	b.WriteString("      apply()\n")
	b.WriteString("      vim.api.nvim_create_autocmd(\"ColorScheme\", {\n")
	b.WriteString("        callback = function()\n")
	b.WriteString("          vim.schedule(apply)\n")
	b.WriteString("        end,\n")
	b.WriteString("      })\n")
	b.WriteString("    end,\n")
	b.WriteString("  },\n")
	b.WriteString("}\n")

	return os.WriteFile(filepath.Join(exportDir, "neovim.lua"), []byte(b.String()), 0644)
}

func writeLuaColor(b *strings.Builder, name, value string) {
	b.WriteString(fmt.Sprintf("        %s = %q,\n", name, value))
}
