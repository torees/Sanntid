with Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Float_Random;
use  Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Float_Random;

procedure ov8 is

    Count_Failed    : exception;    -- Exception to be raised when counting fails
    Gen             : Generator;    -- Random number generator

    protected type Transaction_Manager (N : Positive) is
        entry Finished;
        entry wait_until_aborted;
        function Commit return Boolean;
        procedure Signal_Abort;
    private
    --    wait_until_aborted_Gate_Open : Boolean := False;
        Finished_Gate_Open  : Boolean := False;
        Aborted             : Boolean := False;
        Should_Commit       : Boolean := True;
    end Transaction_Manager;

    protected body Transaction_Manager is
        
        entry wait_until_aborted when Aborted = True is
        begin
        
            if wait_until_aborted'Count = 0 then
                    Aborted := False; 
            end if;
        end wait_until_aborted;


        entry Finished when Finished_Gate_Open or Finished'Count = N is
        begin
            if Finished_Gate_Open = False then
                Finished_Gate_Open := True; 
            end if;           

            
            if Finished'Count = 0 then
                Finished_Gate_Open := False;
                Put_Line("Round complete");
            end if;
        end Finished;


        procedure Signal_Abort is
        begin
            Aborted := True;
        end Signal_Abort;

        function Commit return Boolean is
        begin
            return Should_Commit;
        end Commit;
        
    end Transaction_Manager;

        -------------------------------------------
        -- PART 1: Create the transaction work here
        -------------------------------------------
    
   
    function Unreliable_Slow_Add (x : Integer) return Integer is
    Error_Rate : Constant := 0.15;  -- (between 0 and 1)
    error : Float := Random(Gen);
    
    begin
        if Error_Rate >= error then
            delay Duration(0.5);
            raise Count_Failed;
        else
            delay Duration(4);
            return x +10;
        end if;

    end Unreliable_Slow_Add;




    task type Transaction_Worker (Initial : Integer; Manager : access Transaction_Manager);
    task body Transaction_Worker is
        Num         : Integer   := Initial;
        Prev        : Integer   := Num;
        Round_Num   : Integer   := 0;
    begin
        Put_Line ("Worker" & Integer'Image(Initial) & " started");

        loop
            begin
             Round_Num := Round_Num + 1;
             select
                Manager.wait_until_aborted;
                Num := Num+5;
                Put_Line ("  Worker" & Integer'Image(Initial) & " comitting" & Integer'Image(Num));

            then abort
                
                begin
                Put_Line ("Worker" & Integer'Image(Initial) & " started round" & Integer'Image(Round_Num));
                Num := Unreliable_Slow_Add(Num);
                exception 
                    when Count_Failed =>
                        Manager.Signal_Abort;
                end;
                Manager.Finished;
                Put_Line ("  Worker" & Integer'Image(Initial) & " comitting" & Integer'Image(Num));
        
            end select;

            end;

            delay 0.5;

        end loop;
    end Transaction_Worker;

    Manager : aliased Transaction_Manager (3);

    Worker_1 : Transaction_Worker (0, Manager'Access);
    Worker_2 : Transaction_Worker (1, Manager'Access);
    Worker_3 : Transaction_Worker (2, Manager'Access);

begin
    Reset(Gen); -- Seed the random number generator
end ov8;


