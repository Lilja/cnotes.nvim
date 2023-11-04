local M = {
	opts = {
		localFileDirectory = os.getenv("HOME") .. "/Documents/journal",
		destination = "/config",
		syncBinary = "./sftp-client/cnotes-sftp-client",
		sshHost = "erk-temp",
		fileExtension = ".md",
	},
}
local pluginName = "cnotes"

function M.print(str)
	print(pluginName .. ": " .. str)
end

-- Function to convert relative time string to actual date
function parseRelativeTime(relativeTime)
	local _, _, num, unit = string.find(relativeTime, "(%d+) (%a+)")
	local timeTable = os.date("*t")
	print(relativeTime)

	if relativeTime == "yesterday" then
		timeTable.day = timeTable.day - 1
	elseif relativeTime == "today" then
		-- Do nothing, using current date
	elseif relativeTime == "tomorrow" then
		timeTable.day = timeTable.day + 1
	else
		if unit == "days" then
			timeTable.day = timeTable.day - tonumber(num)
		elseif unit == "months" then
			timeTable.month = timeTable.month - tonumber(num)
		elseif unit == "years" then
			timeTable.year = timeTable.year - tonumber(num)
		else
			print("Invalid time unit")
			return nil
		end
	end
	-- TODO: How can we fix this type hint?
	return os.time(timeTable)
end

function commonDateParser(str)
	local ts
	if str == nil then
		ts = parseRelativeTime("today")
	else
		ts = parseRelativeTime(str)
	end
	if ts == nil then
		M.print("Invalid time range for str '" + str + "'")
		return
	end
	local ymd = os.date("%Y-%m-%d", ts)
	if not exists(M.opts.localFileDirectory) then
		cmd = "mkdir -p " .. M.opts.localFileDirectory
		error(
			"The target directory "
				.. M.opts.localFileDirectory
				.. " does not exist. Create it. Here's how you do it: "
				.. cmd
		)
	end
	local filename = M.opts.localFileDirectory .. "/" .. ymd .. M.opts.fileExtension
	vim.cmd("e " .. filename)
end

function runSyncProcess(filename, force)
	local cmd = M.opts.syncBinary
		.. " -file="
		.. filename
		.. " -dest="
		.. M.opts.destination
		.. " -ssh-host="
		.. M.opts.sshHost

	if force then
		cmd = cmd .. " -force=true"
	end
	cmd = cmd .. " 2>&1"
	local handle = io.popen(cmd)
	if handle ~= nil then
		local result = handle:read("*a")
		handle:close()
		return result
	end
	return ""
end

function handleOutput(output)
	if output:find("OK:") then
	-- do nothing. Don't polute stdout
	elseif output:find("File exists on remote server") then
		M.print("File already exists remotely. Use :JournalForceSync to overwrite")
	else
		M.print(output)
	end
end

function exists(file)
	local ok, err, code = os.rename(file, file)
	if not ok then
		if code == 13 then
			-- Permission denied, but it exists
			return true
		end
	end
	return ok, err
end

function M.setup(opt)
	M.opts = opt
end

function M.forceResync()
	local file = vim.api.nvim_buf_get_name(0)
	if not file:find(M.opts.localFileDirectory) then
		M.print(pluginName .. " not in localFileDirectory( " .. M.opts.localFileDirectory .. ")")
	end
	handleOutput(runSyncProcess(file, true))
end

function M.sync()
	local file = vim.api.nvim_buf_get_name(0)
	if not file:find(M.opts.localFileDirectory) then
		M.print(pluginName .. " not in localFileDirectory( " .. M.opts.localFileDirectory .. ")")
	end
	handleOutput(runSyncProcess(file, false))
end



function M.startJournaling(args)
	commonDateParser(args)
end

return M
